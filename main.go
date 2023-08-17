package main

import (
	"fmt"
	"github.com/ledongthuc/pdf"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
)

type product struct {
	description string
	articleID   int
	price       float64
	unitWeight  float64
	unit        string
	totalPrice  float64
}

func main() {
	f, r, err := pdf.Open("./ica_receipt.pdf")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()

	var (
		result []product
		// Takes a string "LF Lätt CF 13%731869007567716.501 st16.50"
		// And groups them into (LF Lätt CF 13%) (7318690075677) (16.50) (1) (st) (16.50)
		rowRegex = regexp.MustCompile(`^(.*?)(\d{13})(\d+\.\d{2})(\d*\.?\d+)\s(.*?)(\d+\.\d{2})$`)
	)
	for i := 1; i <= r.NumPage(); i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}

		for _, row := range extractTextRows(p) {
			for _, match := range rowRegex.FindAllStringSubmatch(row, -1) {
				if len(match) == 7 {
					articleID, _ := strconv.Atoi(match[2])
					price, _ := strconv.ParseFloat(match[3], 64)
					unitWeight, _ := strconv.ParseFloat(match[4], 64)
					totalPrice, _ := strconv.ParseFloat(match[6], 64)

					result = append(result, product{
						description: match[1],
						articleID:   articleID,
						price:       price,
						unitWeight:  unitWeight,
						unit:        match[5],
						totalPrice:  totalPrice,
					})
				}
			}
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight|tabwriter.Debug)
	defer w.Flush()
	fmt.Fprintln(w, "\tDescription\tArticleID\tPrice\tUnitWeight\tUnit\tTotal price\t")
	fmt.Fprintf(w, "\t\t\t\t\t\t\t\n")
	for _, row := range result {
		fmt.Fprintf(w, "\t%s\t%d\t%.2f\t%.3f\t%s\t%.2f\t\n", row.description, row.articleID, row.price, row.unitWeight, row.unit, row.totalPrice)
	}
}

func extractTextRows(p pdf.Page) []string {
	var (
		// Sometimes the pdf row is not correct and puts the total price on the next row.
		// So if that happens we carry the temporary row that is being created.
		carry  string
		result []string
	)

	rows, err := p.GetTextByRow()
	if err != nil {
		return nil
	}
	for _, row := range rows {
		tmp := carry
		for _, char := range row.Content {
			tmp += char.S
		}

		// If the temporary row ends with either st or kg, we know that something is wrong, because
		// it should have the total price at the end and not the unit. So we carry over to the next row without parsing.
		if strings.HasSuffix(tmp, "st") || strings.HasSuffix(tmp, "kg") {
			carry = tmp
			continue
		}

		result = append(result, tmp)
		carry = ""
	}
	return result
}
