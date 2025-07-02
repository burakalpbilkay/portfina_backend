package handlers

import (
	"database/sql"
	utils "go-airflow-trigger/Utils"
	"log"
	"net/http"
)

func GetBondsHandler(w http.ResponseWriter, r *http.Request) {

	db, err := utils.Connect()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "Database connection failed"})
		return
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT id, type, coupon_rate, currency, buy_sell, discount_curve, maturity_date, term, notional, clean_price,
		       issue_date, inflation_curve, inflation_expectation_curve, underlying_curve_index, last_reset_rate,
		       interest_rate, theoretical_model, compounding, accrual_day_count_convention, day_count_convention,
		       calendar, indexation, created_at
		FROM bonds
	`)
	if err != nil {
		log.Println("DB query error:", err)
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to query bonds"})
		return
	}
	defer rows.Close()

	type Row []interface{}
	var data []Row

	for rows.Next() {
		var (
			id, bondType, currency, buySell, discountCurve, term, inflationCurve,
			inflationExpectationCurve, underlyingCurveIndex, theoreticalModel,
			compounding, accrualConvention, dayCountConvention, calendar, indexation sql.NullString

			couponRate, notional, cleanPrice, lastResetRate, interestRate sql.NullFloat64

			maturityDate, issueDate, createdAt sql.NullTime
		)

		err := rows.Scan(
			&id, &bondType, &couponRate, &currency, &buySell, &discountCurve,
			&maturityDate, &term, &notional, &cleanPrice,
			&issueDate, &inflationCurve, &inflationExpectationCurve, &underlyingCurveIndex,
			&lastResetRate, &interestRate, &theoreticalModel, &compounding,
			&accrualConvention, &dayCountConvention, &calendar, &indexation, &createdAt,
		)
		if err != nil {
			log.Println("Row scan error:", err)
			continue
		}

		row := Row{
			id.String, bondType.String, couponRate.Float64, currency.String, buySell.String,
			discountCurve.String, formatDate(maturityDate), term.String, notional.Float64, cleanPrice.Float64,
			formatDate(issueDate), inflationCurve.String, inflationExpectationCurve.String, underlyingCurveIndex.String,
			lastResetRate.Float64, interestRate.Float64, theoreticalModel.String, compounding.String,
			accrualConvention.String, dayCountConvention.String, calendar.String, indexation.String,
			formatDate(createdAt),
		}
		data = append(data, row)
	}

	columns := []string{
		"id", "type", "coupon_rate", "currency", "buy_sell",
		"discount_curve", "maturity_date", "term", "notional", "clean_price",
		"issue_date", "inflation_curve", "inflation_expectation_curve", "underlying_curve_index",
		"last_reset_rate", "interest_rate", "theoretical_model", "compounding",
		"accrual_day_count_convention", "day_count_convention", "calendar", "indexation",
		"created_at",
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"columns": columns,
		"rows":    data,
	})
}

func formatDate(d sql.NullTime) interface{} {
	if d.Valid {
		return d.Time.Format("2006-01-02")
	}
	return nil
}
