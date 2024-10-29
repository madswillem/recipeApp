package test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/madswillem/recipeApp/internal/tools"
)

func TestCalculateAverage(t *testing.T) {
	t.Run("average positive numbers", func(t *testing.T) {
		avg := tools.CalculateAverage([]float64{1, 2, 3, 4, 5})
		if avg != 3 {
			t.Errorf("Expected average of 3, got %f", avg)
		}
	})
	t.Run("average of positive floats", func(t *testing.T) {
		avg := tools.CalculateAverage([]float64{1.1, 2.2, 3.3, 4.4, 5.5})
		if avg != 3.3 {
			t.Errorf("Expected average of 3.3, got %f", avg)
		}
	})
	t.Run("average of negative numbers", func(t *testing.T) {
		avg := tools.CalculateAverage([]float64{-1, -2, -3, -4, -5})
		if avg != -3 {
			t.Errorf("Expected average of -3, got %f", avg)
		}
	})
	t.Run("average of negative floats", func(t *testing.T) {
		avg := tools.CalculateAverage([]float64{-1.1, -2.2, -3.3, -4.4, -5.5})
		if avg != -3.3 {
			t.Errorf("Expected average of -3.3, got %f", avg)
		}
	})
	t.Run("average of negative and positive numbers", func(t *testing.T) {
		avg := tools.CalculateAverage([]float64{-1, 2, -3, 4, -5})
		if avg != -0.6 {
			t.Errorf("Expected average of 0, got %f", avg)
		}
	})
}

func TestGetCurrentData(t *testing.T) {
	_, err := tools.GetCurrentData()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestAverageVectors(t *testing.T) {
	v1 := []float64{1.0, 2.0, 3.0}
	v2 := []float64{4.0, 5.0}
	v3 := []float64{7.0, 8.0, 9.0, 10.0}

	expected := []float64{4, 5, 4, 3.333333333333333}

	avg := tools.AverageVectors(v1, v2, v3)

	fmt.Println(avg)
	if !reflect.DeepEqual(avg, expected) {
		t.Errorf("Expected %v but got %v", expected, avg)
	}
}

func TestAddVectors(t *testing.T) {
	t.Run("Test with 3 Vecs of diffrent length", func(t *testing.T) {
		vec1 := []float64{1.0, 2.0, 3.0}
		vec2 := []float64{4.0, 5.0}
		vec3 := []float64{7.0, 8.0, 9.0, 10.0}

		sum := tools.AddVectors(vec1, vec2, vec3)

		fmt.Println("Sum of vectors:", sum)
	})
}

func TestMultiplyVectorByNum(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		vec := []float64{1.0, 3.0, 2.0}

		product := tools.MultiplyVectorByNum(10, vec)

		fmt.Println("Product: ", product)
	})
}

func TestMergeMatrix(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		a := tools.Matrix{
			Len: 1,
			Dict: map[string]int{
				"hi":  1,
				"i":   2,
				"am":  3,
				"ben": 4,
			},
			Vec: []float64{
				1, 1, 1, 1,
			},
		}
		b := tools.Matrix{
			Len: 1,
			Dict: map[string]int{
				"hi":      1,
				"you":     2,
				"are":     3,
				"timothe": 4,
			},
			Vec: []float64{
				1, 1, 1, 1,
			},
		}

		expected := tools.Matrix{
			Dict: map[string]int{
				"hi":      1,
				"i":       2,
				"am":      3,
				"ben":     4,
				"you":     5,
				"are":     6,
				"timothe": 7,
			},
			Vec: []float64{
				1, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5,
			},
		}

		r := tools.MergeMatrix(a, b)

		if !reflect.DeepEqual(r.Dict, expected.Dict) {
			t.Errorf("Expected %+v but got %+v", expected.Dict, r.Dict)
		}
		if !reflect.DeepEqual(r.Vec, expected.Vec) {
			t.Errorf("Expected %+v but got %+v", expected.Vec, r.Vec)
		}
	})
}
