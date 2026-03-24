package domain


import "time"

type Balance struct {
	UserID    string    
	Total     float64   
	Currency  string    
	UpdatedAt time.Time 
}

type CategoryStats struct {
	Category string 
	Amount   float64 
	Count    int     
}

type Statistics struct {
	Balance           Balance         
	Income            float64         
	Expenses          float64         
	NetFlow           float64         
	CategoryBreakdown []*CategoryStats 
	PeriodStart       time.Time       
	PeriodEnd         time.Time       
}
