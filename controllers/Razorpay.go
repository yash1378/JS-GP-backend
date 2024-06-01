package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/razorpay/razorpay-go/utils"
)

type order struct {
	Amount int `json:"amount"`
}

type veri struct {
	Payid   string `json:"razorpay_payment_id"`
	Orderid string `json:"razorpay_order_id"`
	Sign    string `json:"razorpay_signature"`
}

// func Order(c *gin.Context) {
// 	var input order
// 	fmt.Println(c.Request.Body)
// 	if err := c.ShouldBindBodyWithJSON(&input); err != nil {
// 		fmt.Println(input)
// 		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// 	// println(input)
// 	// amt, err := strconv.Atoi(input.Amount)
// 	client := razorpay.NewClient("rzp_test_gAnipV9pamMXV2", "cUhcdokPA3QwiULcMmadGx86")

// 	data := map[string]interface{}{
// 		"amount":   input.Amount * 100, // amount in paise, so 50000 paise = INR 500
// 		"currency": "INR",
// 		"receipt":  "some_receipt_id",
// 	}
// 	body, err := client.Order.Create(data, nil)
// 	if err != nil {
// 		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": body})
// }

func Order(c *gin.Context) {
	var input map[string]interface{}

	if err := c.ShouldBindJSON(&input); err != nil {
		fmt.Println(input)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Print the input received from the client
	fmt.Println(input)

	// Return the input in the response
	c.JSON(http.StatusOK, gin.H{"message": input})
}

func Key(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"key": "rzp_test_gAnipV9pamMXV2"})
}

func Verify(c *gin.Context) {
	var input veri

	if err := c.ShouldBindBodyWithJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// client := razorpay.NewClient("rzp_test_gAnipV9pamMXV2", "cUhcdokPA3QwiULcMmadGx86")

	params := map[string]interface{}{
		"razorpay_order_id":   input.Orderid,
		"razorpay_payment_id": input.Payid,
	}

	signature := input.Sign
	secret := "cUhcdokPA3QwiULcMmadGx86"
	utils.VerifyPaymentSignature(params, signature, secret)

	redirectURL := fmt.Sprintf("http://localhost:3000/razor?reference=%v", input.Payid)
	c.Redirect(http.StatusFound, redirectURL)

	// c.JSON(http.StatusOK, gin.H{"message": "payment success"})
}
