package util

import (
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
)

func zeroTrio(err error) (pgtype.Numeric, float64, error) {
	return pgtype.Numeric{Valid: false}, 0, err
}

// NumericToFloat64 ...
func NumericToFloat64(num *pgtype.Numeric) (float64, error) {
	f64value, err := num.Float64Value()
	if err != nil {
		return 0, err
	}
	if !f64value.Valid {
		return 0, fmt.Errorf("unable to convert %+v to float64, got: %+v", num, f64value)
	}
	f64 := f64value.Float64
	return f64, nil
}

func Float64ToNumeric(n float64) (pgtype.Numeric, error) {
	nStr := strconv.FormatFloat(n, 'f', -1, 64)
	numeric := pgtype.Numeric{}
	err := numeric.Scan(nStr)
	if err != nil {
		return numeric, err
	}
	return numeric, nil
}

func ChangeSignNumeric(num *pgtype.Numeric) (newNum pgtype.Numeric, err error) {
	n, err := NumericToFloat64(num)
	if err != nil {
		return newNum, err
	}

	newNum, err = Float64ToNumeric(-n)
	if err != nil {
		return newNum, err
	}

	return newNum, nil
}

func AddFloat64ToNumeric(num *pgtype.Numeric, delta float64) (pgtype.Numeric, float64, error) {
	curr, err := NumericToFloat64(num)
	if err != nil {
		return zeroTrio(err)
	}
	curr += delta

	newNum, err := Float64ToNumeric(curr)
	if err != nil {
		return zeroTrio(err)
	}

	return newNum, curr, nil
}

func AddNumericToNumeric(a *pgtype.Numeric, b *pgtype.Numeric) (pgtype.Numeric, float64, error) {
	aFloat, err := NumericToFloat64(a)
	if err != nil {
		return zeroTrio(err)
	}
	bFloat, err := NumericToFloat64(b)
	if err != nil {
		return zeroTrio(err)
	}

	newNumFloat := aFloat + bFloat
	newNum, err := Float64ToNumeric(newNumFloat)
	if err != nil {
		return zeroTrio(err)
	}

	return newNum, newNumFloat, nil
}

func DeltaOfTwoNumeric(a *pgtype.Numeric, b *pgtype.Numeric) (delta float64, err error) {
	aFloat, err := NumericToFloat64(a)
	if err != nil {
		return delta, err
	}
	bFloat, err := NumericToFloat64(b)
	if err != nil {
		return delta, err
	}

	delta = aFloat - bFloat
	return delta, nil
}

// if sign of b is not provided, it defaults to 1 (positive)
func AddNumToNum(a any, b any, args ...int) (pgtype.Numeric, float64, error) {
	sign := 1
	if len(args) > 0 && args[0] == -1 {
		sign = -1
	}

	aFloat, err := extractNumericAsFloat(a)
	if err != nil {
		return zeroTrio(err)
	}
	bFloat, err := extractNumericAsFloat(b)
	if err != nil {
		return zeroTrio(err)
	}

	newNumFloat := aFloat
	if sign == 1 {
		newNumFloat += bFloat
	} else {
		newNumFloat -= bFloat
	}
	newNum, err := Float64ToNumeric(newNumFloat)
	if err != nil {
		return zeroTrio(err)
	}
	return newNum, newNumFloat, nil
}

func extractNumericAsFloat(a any) (float64, error) {
	switch a := a.(type) {
	case pgtype.Numeric:
		return NumericToFloat64(&a)
	case *pgtype.Numeric:
		return NumericToFloat64(a)
	case float64:
		return a, nil
	default:
		return 0, fmt.Errorf("unsupported type %T", a)
	}
}
