// Package rts implements Real Time Statistics calculations module
package rts

import "time"

/**
 * referrers = {
 * 		'url_1': N,
 * 		'url_2': N
 * }
 *
 * visits = {
 *		'Yesterday': 'DAY_2',
 *		'Today': 'DAY_1',
 *		'Data' : {
 * 			'DAY_1': {sessionID_1: timestamp, sessionID_2: timestamp, .., sessionID_N: timestamp},
 *			'...': 	 {sessionID_1: timestamp, sessionID_2: timestamp, .., sessionID_N: timestamp},
 *			'DAY_N': {sessionID_1: timestamp, sessionID_2: timestamp, .., sessionID_N: timestamp},
 *		}
 * }
 *
 * conversions = {
 *		"addedToCart": {sessionID_1: true, sessionID_2: true, .., sessionID_N: true},
 *		"reachedCheckout": {sessionID_1: true, sessionID_2: true, .., sessionID_N: true},
 *		"purchased": {sessionID_1: true, sessionID_2: true, .., sessionID_N: true},
 *		"visitors": X
 * }
 *
 * sales = {
 *		"lastUpdate": timestamp,
 *		"today": x,
 *		"yesterday": y,
 *		"ratio": z,
 * }
 *
 * salesDetail = {
 *		"period(MD5(dateFrom/dateTo))": {
 *			"Data": {},
 *			"lastUpdate": timestamp
 *		}
 * }
 *
 * topSellers = {
 *		"lastUpdate": timestamp,
 *		"Data": {
 *			"itemID_1": {
 *				"Name": "XXX",
 *				"Image": "YYY",
 *				"Count": X,
 *			},
 *			...
 *			"itemID_N": {
 *				"Name": "XXX",
 *				"Image": "YYY",
 *				"Count": X,
 *			},
 *		},
 * }
 *
 * type OnlineReferer struct {
 * 		type int
 * 		time int
 * }
 *
 * OnlineSessions map[string]OnlineReferer
 * OnlineDirect int
 * OnlineSite int
 * OnlineSearch int
 *
 *
 */

// Package global constants
const (
	ConstCollectionNameRTSSalesHistory = "rts_sales_history"
	ConstCollectionNameRTSSales        = "rts_sales"
	ConstCollectionNameRTSVisitors     = "rts_visitors"

	ConstReferrerTypeDirect = 0
	ConstReferrerTypeSite   = 1
	ConstReferrerTypeSearch = 2

	ConstVisitorAddToCart     = 1
	ConstVisitorCheckout      = 2
	ConstVisitorSales         = 3
	ConstVisitorOnlineSeconds = 10
)

// Package global variables
var (
	referrers             = make(map[string]int)
	visitorsInfoToday     = new(dbVisitorRow)
	visitorsInfoYesterday = new(dbVisitorRow)

	sales       = Sales{}
	salesDetail = make(map[string]*SalesDetailData)
	topSellers  = new(TopSellers)

	OnlineSessions    = make(map[string]*OnlineReferrer)
	OnlineDirect      = 0
	OnlineSite        = 0
	OnlineSearch      = 0
	OnlineSessionsMax = 0
	OnlineDirectMax   = 0
	OnlineSiteMax     = 0
	OnlineSearchMax   = 0

	searchEngines = []string{"www.daum.net", "www.google.com", "www.eniro.se", "www.naver.com", "www.yahoo.com",
		"www.msn.com", "www.bing.com", "www.aol.com", "www.aol.com", "www.lycos.com", "www.ask.com", "www.altavista.com",
		"search.netscape.com", "www.cnn.com", "www.about.com", "www.mamma.com", "www.alltheweb.com", "www.voila.fr",
		"search.virgilio.it", "www.bing.com", "www.baidu.com", "www.alice.com", "www.yandex.com", "www.najdi.org.mk",
		"www.aol.com", "www.mamma.com", "www.seznam.cz", "www.search.com", "www.wp.pl", "online.onetcenter.org",
		"www.szukacz.pl", "www.yam.com", "www.pchome.com", "www.kvasir.no", "sesam.no", "www.ozu.es", "www.terra.com",
		"www.mynet.com", "www.ekolay.net", "www.rambler.ru"}
)

// Visits - unknown purpose structure
type Visits struct {
	Data      map[string]map[string]int32
	Yesterday string
	Today     string
}

// VisitorDetail - unknown purpose structure
type VisitorDetail struct {
	Time     time.Time
	Checkout int
}

// dbVisitorRow - unknown purpose structure
type dbVisitorRow struct {
	ID       string
	Day      time.Time
	Visitors int
	Cart     int
	Checkout int
	Sales    int
	Details  map[string]*VisitorDetail
}

// Sales - unknown purpose structure
type Sales struct {
	lastUpdate int64
	today      int
	yesterday  int
	ratio      float64
}

// SalesDetailData - unknown purpose structure
type SalesDetailData struct {
	Data       map[string]int
	lastUpdate int64
}

// TopSellers - unknown purpose structure
type TopSellers struct {
	Data       map[string]*SellerInfo
	lastUpdate int64
}

// SellerInfo - unknown purpose structure
type SellerInfo struct {
	Name  string
	Image string
	Count int
}

// OnlineReferrer - unknown purpose structure
type OnlineReferrer struct {
	referrerType int
	time         time.Time
}
