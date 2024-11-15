package cart

// todo wrong calculation
// calculate order tax amount
func CalculateItemsSalesTax(items []CartItem) []CartItem {
	for i := 0; i < len(items); i++ {
		item := &items[i]
		item.PriceWithTax = 0
		item.TotalPrice = 0
		item.TotalTax = 0

		price := item.PurchasePrice
		quantity := float64(item.Quantity)
		item.TotalPrice = float64(price * quantity)

	}

	return items
}
