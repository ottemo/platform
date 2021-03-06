package reporting

import (
	"time"
)

// Package global constants
const (

	ConstErrorModule = "reporting"
	ConstErrorLevel  = 6
)

// ProductPerfItem is a container for sales by item reporting
type ProductPerfItem struct {
	Name       string  `json:"name"`
	Sku        string  `json:"sku"`
	GrossSales float64 `json:"gross_sales"`
	UnitsSold  int     `json:"units_sold"`
}

// ProductPerf is an array of sales by item structs, ProductPerfItem, to be sorted
type ProductPerf []ProductPerfItem

func (a ProductPerf) Len() int {
	return len(a)
}

func (a ProductPerf) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ProductPerf) Less(i, j int) bool {
	if a[i].UnitsSold == a[j].UnitsSold {
		return a[i].Name < a[j].Name
	}

	return a[i].UnitsSold > a[j].UnitsSold
}

// CustomerActivityItem is a container for visitor stats over time
type CustomerActivityItem struct {
	Email            string    `json:"email"`
	Name             string    `json:"name"`
	TotalSales       float64   `json:"total_sales"`
	TotalOrders      int       `json:"total_orders"`
	AverageSales     float64   `json:"avg_sales"`
	EarliestPurchase time.Time `json:"earliest_purchase"`
	LatestPurchase   time.Time `json:"latest_purchase"`
}

// CustomerActivityBySales is an array of CustomerActivityItems to be sorted by total sales.
type CustomerActivityBySales []CustomerActivityItem

func (a CustomerActivityBySales) Len() int {
	return len(a)
}

func (a CustomerActivityBySales) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a CustomerActivityBySales) Less(i, j int) bool {
	if a[i].TotalSales == a[j].TotalSales {
		return a[i].Email < a[j].Email
	}

	return a[i].TotalSales > a[j].TotalSales
}

// CustomerActivityByOrders is a array of CustomerActivityItems to be sorted by order count.
type CustomerActivityByOrders []CustomerActivityItem

func (a CustomerActivityByOrders) Len() int {
	return len(a)
}

func (a CustomerActivityByOrders) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a CustomerActivityByOrders) Less(i, j int) bool {
	if a[i].TotalOrders == a[j].TotalOrders {
		return a[i].Email < a[j].Email
	}

	return a[i].TotalOrders > a[j].TotalOrders
}

// StatItem is a container for sales data by product sku over time.
type StatItem struct {
	Key          string  `json:"key"`
	Name         string  `json:"name"`
	TotalSales   float64 `json:"total_sales"`
	TotalOrders  int     `json:"total_orders"`
	AverageSales float64 `json:"avg_sales"`
}

// StatsBySales is a array to be sorted of StatItems by sales totals over time.
type StatsBySales []StatItem

func (a StatsBySales) Len() int {
	return len(a)
}

func (a StatsBySales) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a StatsBySales) Less(i, j int) bool {
	if a[i].TotalSales == a[j].TotalSales {
		return a[i].Name < a[j].Name
	}

	return a[i].TotalSales > a[j].TotalSales
}
