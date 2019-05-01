package mssql

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/ottemo/commerce/db"
	"github.com/ottemo/commerce/env"
	"github.com/ottemo/commerce/utils"
)

// LoadByID capture all records from the DB based on id
func (it *DBCollection) LoadByID(id string) (map[string]interface{}, error) {
	var result map[string]interface{}

	if !ConstUseUUIDids {
		if err := it.AddFilter("_id", "=", id); err != nil {
			return result, env.ErrorDispatch(err)
		}
	} else {
		if idValue, err := strconv.ParseInt(id, 10, 64); err == nil {
			if err := it.AddFilter("_id", "=", idValue); err != nil {
				return result, env.ErrorDispatch(err)
			}
		} else {
			if err := it.AddFilter("_id", "=", id); err != nil {
				return result, env.ErrorDispatch(err)
			}
		}
	}

	err := it.Iterate(func(row map[string]interface{}) bool {
		result = row
		return false
	})

	if len(result) == 0 {
		err = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "96aa4214-3c7e-40df-934c-e0584e21dc95", "not found")
	}
	return result, err
}

// Load returns all records from DB for the current collection and apply a filter if set
func (it *DBCollection) Load() ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	err := it.Iterate(func(row map[string]interface{}) bool {
		result = append(result, row)
		return true
	})

	return result, env.ErrorDispatch(err)
}

// Iterate applies [iterator] function to each record, stops on return false
func (it *DBCollection) Iterate(iteratorFunc func(record map[string]interface{}) bool) error {

	SQL := it.getSelectSQL()

	rows, err := connectionQuery(SQL)
	defer closeCursor(rows)

	if err == nil {
		for ok := rows.Next(); ok == true; ok = rows.Next() {
			if row, err := getRowAsStringMap(rows); err == nil {
				it.modifyResultRow(row)

				if !iteratorFunc(row) {
					break
				}
			}
		}
	}

	if err == io.EOF {
		err = nil
	} else if err != nil {
		err = sqlError(SQL, err)
	}

	return env.ErrorDispatch(err)
}

// Distinct returns distinct values of specified attribute
func (it *DBCollection) Distinct(columnName string) ([]interface{}, error) {

	prevResultColumns := it.ResultColumns
	if err := it.SetResultColumns(columnName); err != nil {
		_ = env.ErrorDispatch(err)
	}

	limit := " "
	if it.Limit > 0 {
		limit = " TOP " + utils.InterfaceToString(it.Limit) + " "
	}
	SQL := "SELECT" + limit + "DISTINCT " + it.getSQLResultColumns() + " FROM [" + it.Name + "]" + it.getSQLFilters() + it.getSQLOrder()

	it.ResultColumns = prevResultColumns

	rows, err := connectionQuery(SQL)
	defer closeCursor(rows)

	var result []interface{}
	if err == nil {
		for ok := rows.Next(); ok == true; ok = rows.Next() {
			if row, err := getRowAsStringMap(rows); err == nil {
				ignoreNull := false
				for _, columnValue := range row {
					if columnValue == nil {
						ignoreNull = true
					}
				}
				if ignoreNull {
					continue
				}

				it.modifyResultRow(row)

				for _, columnValue := range row {
					// if value is array then we need to make distinct within array by self
					if arrayValue, ok := columnValue.([]interface{}); ok {
						for _, arrayItem := range arrayValue {
							isAlreadyInResult := false
							// looking for array item value in result array
							for _, resultItem := range result {
								if arrayItem == resultItem {
									isAlreadyInResult = true
									break
								}
							}
							if !isAlreadyInResult {
								result = append(result, arrayItem)
							}
						}
					} else {
						// if value is not array then SQL did distinct work for us
						result = append(result, columnValue)
					}

					break
				}
			}
		}
	}

	if err == io.EOF {
		err = nil
	} else if err != nil {
		err = sqlError(SQL, err)
	}

	return result, env.ErrorDispatch(err)
}

// Count returns count of rows matching current select statement
func (it *DBCollection) Count() (int, error) {
	sqlLoadFilter := it.getSQLFilters()

	SQL := "SELECT COUNT(*) AS cnt FROM [" + it.Name + "]" + sqlLoadFilter

	rows, err := connectionQuery(SQL)
	defer closeCursor(rows)

	if err == nil {
		if row, err := getRowAsStringMap(rows); err == nil {
			cnt := utils.InterfaceToInt(row["cnt"])
			return cnt, err
		}
	}

	if err == io.EOF {
		err = nil
	} else if err != nil {
		err = sqlError(SQL, err)
	}

	return 0, err
}

// Save stores record in DB for current collection
func (it *DBCollection) Save(item map[string]interface{}) (string, error) {

	// prevents saving of blank records
	if len(item) == 0 {
		return "", nil
	}

	// we should make new _id column if it was not set
	if ConstUseUUIDids {
		if idValue, present := item["_id"]; !present || idValue == nil {
			item["_id"] = it.makeUUID("")
		} else {
			if idValue, ok := idValue.(string); ok {
				item["_id"] = it.makeUUID(idValue)
			}
		}
	} else {
		// _id in MySQL supposed to be auto-incremented int but for MongoDB it forced to be string
		// collection interface also forced us to use string but we still want it ti be int in DB
		// to make that we need to convert it before save from  string to int or nil
		// and after save get auto-incremented id as convert to string
		if idValue, present := item["_id"]; present && idValue != nil {
			if idValue, ok := idValue.(string); ok {

				if intValue, err := strconv.ParseInt(idValue, 10, 64); err == nil {
					item["_id"] = intValue
				} else {
					item["_id"] = nil
				}

			} else {
				return "", env.ErrorNew(ConstErrorModule, ConstErrorLevel, "e6bb61d5-c16b-4463-89a8-cc0bf99972a4", "unexpected _id value '"+fmt.Sprint(item)+"'")
			}
		} else {
			item["_id"] = nil
		}
	}

	// SQL generation
	columns := make([]string, 0, len(item))
	args := make([]string, 0, len(item))
	columnEqArg := make([]string, 0, len(item))
	using := make([]string, 0, len(item))

	values := make([]interface{}, 0, len(item))

	for key, value := range item {
		if item[key] != nil {

			sqlValue := convertValueForSQL(value, it.GetColumnType(key))
			using = append(using, sqlValue+" AS ["+key+"]")

			if key == "_id" && !ConstUseUUIDids {
				continue
			}

			columns = append(columns, "["+key+"]")
			args = append(args, sqlValue)

			if key != "_id" {
				columnEqArg = append(columnEqArg, "["+key+"]="+sqlValue)
			}

			//args = append(args, "$_"+key)
			//values = append(values, convertValueForSQL(value))
		}
	}

	SQL := "INSERT INTO [" + it.Name + "] (" + strings.Join(columns, ", ") + ") VALUES (" + strings.Join(args, ",") + ");"
	if item["_id"] != nil {
		SQL = "SELECT COUNT(*) FROM  [" + it.Name + "] WHERE [_id] = " + convertValueForSQL(item["_id"], db.ConstTypeID)
		if rows, err := dbEngine.connection.Query(SQL); err == nil && rows.Next() {
			var cnt int64
			if err := rows.Scan(&cnt); err == nil && cnt > 0 {
				SQL = "UPDATE [" + it.Name + "] SET " + strings.Join(columnEqArg, ", ")
			}

		}
		//SQL = "MERGE [" + it.Name + "] AS t " +
		//" USING (SELECT " + strings.Join(using, ", ") + ") AS s ON (t._id = s._id) " +
		//" WHEN MATCHED THEN " +
		//" UPDATE SET " + strings.Join(columnEqArg, ", ") +
		//" WHEN NOT MATCHED THEN " +
		//" INSERT (" + strings.Join(columns, ", ") + ") VALUES (" + strings.Join(args, ",") + ");";

		//SQL = "IF (SELECT COUNT(*) FROM  [" + it.Name + "] WHERE [_id] = '" + utils.InterfaceToString(item["_id"]) + "') > 0 " +
		//	" UPDATE [" + it.Name + "] SET " + strings.Join(columnEqArg, ", ") +
		//	" ELSE " +
		//	" INSERT INTO [" + it.Name + "] (" + strings.Join(columns, ", ") + ") VALUES (" + strings.Join(args, ",") + ")";
	}

	if SQL[0:6] == "INSERT" && !ConstUseUUIDids {
		newIDInt64, err := connectionExecWLastInsertID(SQL, values...)
		if err != nil {
			return "", sqlError(SQL, err)
		}

		// auto-incremented _id back to string
		newIDString := strconv.FormatInt(newIDInt64, 10)
		item["_id"] = newIDString
	} else {
		err := connectionExec(SQL, values...)
		if err != nil {
			return "", sqlError(SQL, err)
		}
	}

	return utils.InterfaceToString(item["_id"]), nil
}

// Delete removes records that matches current select statement from DB
//   - returns amount of affected rows
func (it *DBCollection) Delete() (int, error) {
	sqlDeleteFilter := it.getSQLFilters()

	SQL := "DELETE FROM [" + it.Name + "] " + sqlDeleteFilter

	affected, err := connectionExecWAffected(SQL)

	return int(affected), env.ErrorDispatch(err)
}

// DeleteByID removes record from DB by is's id
func (it *DBCollection) DeleteByID(id string) error {
	SQL := "DELETE FROM [" + it.Name + "] WHERE [_id] = " + convertValueForSQL(id, db.ConstTypeID)

	return connectionExec(SQL)
}

// SetupFilterGroup setups filter group params for collection
func (it *DBCollection) SetupFilterGroup(groupName string, orSequence bool, parentGroup string) error {
	if _, present := it.FilterGroups[parentGroup]; !present && parentGroup != "" {
		return env.ErrorNew(ConstErrorModule, ConstErrorLevel, "0d7c8917-4502-478f-92b5-1a10dc2036c8", "invalid parent group")
	}

	filterGroup := it.getFilterGroup(groupName)
	filterGroup.OrSequence = orSequence
	filterGroup.ParentGroup = parentGroup

	return nil
}

// RemoveFilterGroup removes filter group for collection
func (it *DBCollection) RemoveFilterGroup(groupName string) error {
	if _, present := it.FilterGroups[groupName]; !present {
		return env.ErrorNew(ConstErrorModule, ConstErrorLevel, "07d09c7d-c98e-460b-9f79-53eefa0ced6c", "invalid group name")
	}

	delete(it.FilterGroups, groupName)
	return nil
}

// AddGroupFilter adds selection filter to specific filter group (all filter groups will be joined before db query)
func (it *DBCollection) AddGroupFilter(groupName string, columnName string, operator string, value interface{}) error {
	err := it.updateFilterGroup(groupName, columnName, operator, value)
	if err != nil {
		return err
	}

	return nil
}

// AddStaticFilter adds selection filter that will not be cleared by ClearFilters() function
func (it *DBCollection) AddStaticFilter(columnName string, operator string, value interface{}) error {

	err := it.updateFilterGroup(ConstFilterGroupStatic, columnName, operator, value)
	if err != nil {
		return err
	}

	return nil
}

// AddFilter adds selection filter to current collection(table) object
func (it *DBCollection) AddFilter(ColumnName string, Operator string, Value interface{}) error {

	err := it.updateFilterGroup(ConstFilterGroupDefault, ColumnName, Operator, Value)
	if err != nil {
		return err
	}

	return nil
}

// ClearFilters removes all filters that were set for current collection, except static
func (it *DBCollection) ClearFilters() error {
	for filterGroup := range it.FilterGroups {
		if filterGroup != ConstFilterGroupStatic {
			delete(it.FilterGroups, filterGroup)
		}
	}

	return nil
}

// AddSort adds sorting for current collection
func (it *DBCollection) AddSort(ColumnName string, Desc bool) error {
	if it.HasColumn(ColumnName) {
		if Desc {
			it.Order = append(it.Order, ColumnName+" DESC")
		} else {
			it.Order = append(it.Order, ColumnName)
		}
	} else {
		return env.ErrorNew(ConstErrorModule, ConstErrorLevel, "57703bc5-d3a5-4367-92e5-960d5582a407", "can't find column '"+ColumnName+"'")
	}

	return nil
}

// ClearSort removes any sorting that was set for current collection
func (it *DBCollection) ClearSort() error {
	it.Order = make([]string, 0)
	return nil
}

// SetResultColumns limits column selection for Load() and LoadByID()function
func (it *DBCollection) SetResultColumns(columns ...string) error {
	for _, columnName := range columns {
		if !it.HasColumn(columnName) {
			return env.ErrorNew(ConstErrorModule, ConstErrorLevel, "82fad6da-5970-42dc-bd29-306b65ef4dbd", "there is no column "+columnName+" found")
		}

		it.ResultColumns = append(it.ResultColumns, columnName)
	}

	return nil
}

// SetLimit results pagination
func (it *DBCollection) SetLimit(Offset int, Limit int) error {
	it.Limit = Limit
	// it.Offset = Offset

	if Offset > 0 {
		return env.ErrorNew(ConstErrorModule, ConstErrorLevel, "6ad9cd8c-b52e-46c8-bd91-174fe71774a1", "offset currently is unsupported ")
	}

	return nil
}

// ListColumns returns attributes(columns) available for current collection(table)
func (it *DBCollection) ListColumns() map[string]string {

	result := make(map[string]string)

	if ConstUseUUIDids {
		result["_id"] = "int"
	} else {
		result["_id"] = "varchar"
	}

	// updating column into collection
	SQL := "SELECT [column], [type] FROM [" + ConstCollectionNameColumnInfo + "] WHERE [collection] = '" + it.Name + "'"
	rows, _ := connectionQuery(SQL)
	defer closeCursor(rows)

	for ok := rows.Next(); ok == true; ok = rows.Next() {
		row, err := getRowAsStringMap(rows)
		if err != nil {
			_ = env.ErrorDispatch(err)
		}

		key := row["column"].(string)
		value := row["type"].(string)

		result[key] = value
	}

	// updating cached attribute types information
	if _, present := dbEngine.attributeTypes[it.Name]; !present {
		dbEngine.attributeTypes[it.Name] = make(map[string]string)
	}

	dbEngine.attributeTypesMutex.Lock()
	for attributeName, attributeType := range result {
		dbEngine.attributeTypes[it.Name][attributeName] = attributeType
	}
	dbEngine.attributeTypesMutex.Unlock()

	return result
}

// GetColumnType returns SQL like type of attribute in current collection, or if not present ""
func (it *DBCollection) GetColumnType(columnName string) string {
	if columnName == "_id" {
		return db.ConstTypeID
	}

	// looking in cache first
	attributeType, present := dbEngine.attributeTypes[it.Name][columnName]
	if !present {
		// updating cache, and looking again
		it.ListColumns()
		attributeType, present = dbEngine.attributeTypes[it.Name][columnName]
	}

	return attributeType
}

// HasColumn checks attribute(column) presence in current collection
func (it *DBCollection) HasColumn(columnName string) bool {
	// looking in cache first
	_, present := dbEngine.attributeTypes[it.Name][columnName]
	if !present {
		// updating cache, and looking again
		it.ListColumns()
		_, present = dbEngine.attributeTypes[it.Name][columnName]
	}

	return present
}

// AddColumn adds new attribute(column) to current collection(table)
func (it *DBCollection) AddColumn(columnName string, columnType string, indexed bool) error {

	// checking column name
	if !ConstSQLNameValidator.MatchString(columnName) {
		return env.ErrorNew(ConstErrorModule, ConstErrorLevel, "528f0656-92e2-4069-aae9-97323ef43094", "not valid column name for DB engine: "+columnName)
	}

	// checking if column already present
	if it.HasColumn(columnName) {
		if currentType := it.GetColumnType(columnName); currentType != columnType {
			return env.ErrorNew(ConstErrorModule, ConstErrorLevel, "5d10721d-9ab5-4484-a142-052cfc7c8c41", "column '"+columnName+"' already exists with type '"+currentType+"' for '"+it.Name+"' collection. Requested type '"+columnType+"'")
		}
		return nil
	}

	// updating collection info table
	//--------------------------------
	SQL := "INSERT INTO [" + ConstCollectionNameColumnInfo + "] ([collection], [column], [type], [indexed]) VALUES (" +
		"'" + it.Name + "', " +
		"'" + columnName + "', " +
		"'" + columnType + "', "
	if indexed {
		SQL += "1)"
	} else {
		SQL += "0)"
	}

	err := connectionExec(SQL)
	if err != nil {
		return sqlError(SQL, err)
	}

	// updating physical table
	//-------------------------
	ColumnType, err := GetDBType(columnType)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	SQL = "ALTER TABLE [" + it.Name + "] ADD [" + columnName + "] " + ColumnType

	err = connectionExec(SQL)
	if err != nil {
		return sqlError(SQL, err)
	}

	// updating collection columns list
	it.ListColumns()

	return nil
}

// RemoveColumn removes attribute(column) to current collection(table)
func (it *DBCollection) RemoveColumn(columnName string) error {

	// checking column in table
	//-------------------------
	if columnName == "_id" {
		return env.ErrorNew(ConstErrorModule, ConstErrorLevel, "d444e354-0f1d-4781-acce-3e746c6ecc66", "you can't remove _id column")
	}

	if !it.HasColumn(columnName) {
		return env.ErrorNew(ConstErrorModule, ConstErrorLevel, "131311b3-8f7c-4c8f-a05f-231d1c6d3c80", "column '"+columnName+"' not exists in '"+it.Name+"' collection")
	}

	SQL := "DELETE FROM [" + ConstCollectionNameColumnInfo + "] WHERE [collection]='" + it.Name + "' AND [column]='" + columnName + "'"
	if err := connectionExec(SQL); err != nil {
		return sqlError(SQL, err)
	}

	// updating physical table
	//-------------------------
	SQL = "ALTER TABLE [" + it.Name + "] DROP COLUMN [" + columnName + "] "

	err := connectionExec(SQL)
	if err != nil {
		return sqlError(SQL, err)
	}

	if _, present := dbEngine.attributeTypes[it.Name]; present {
		if _, present = dbEngine.attributeTypes[it.Name][columnName]; present {
			delete(dbEngine.attributeTypes[it.Name], columnName)
		}
	}

	it.ListColumns()

	return nil
}
