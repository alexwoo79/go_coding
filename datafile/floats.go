package datafile

import (
	"bufio"
	"os"
	"strconv"

	expr "github.com/Knetic/govaluate"
)

// ReadFile reads a file and returns its contents as a slice of float64
func GetFloats(filename string) ([]float64, error) {
	var numbers []float64
	var number float64
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// 先尝试直接解析为数字
		number, err = strconv.ParseFloat(line, 64)
		if err == nil {
			numbers = append(numbers, number)
			continue
		}

		// 如果不是纯数字，则尝试作为表达式解析
		expression, evalErr := expr.NewEvaluableExpression(line)
		if evalErr != nil {
			return nil, evalErr
		}

		result, evalErr := expression.Evaluate(nil)
		if evalErr != nil {
			return nil, evalErr
		}

		switch v := result.(type) {
		case float64:
			numbers = append(numbers, v)
		case int:
			numbers = append(numbers, float64(v))
		default:
			return nil, strconv.ErrSyntax
		}
	}

	err = file.Close()
	if err != nil {
		return nil, err
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	return numbers, nil
}
