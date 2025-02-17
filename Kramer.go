package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time" // Добавляем импорт пакета "time" для работы со временем
)

func main() {
	filename := "matrix1.txt"

	// Чтение матрицы из файла
	matrix, err := readMatrixFromFile(filename)
	if err != nil {
		fmt.Println("Ошибка чтения матрицы из файла:", err)
		return
	}

	// Проверка, является ли матрица расширенной матрицей для метода Крамера (квадратная + 1 столбец)
	if !isAugmentedMatrixForCramer(matrix) {
		fmt.Println("Матрица не соответствует формату для метода Крамера.")
		return
	}

	// Отделение матрицы коэффициентов от столбца свободных членов
	coefficientMatrix := make([][]float64, len(matrix))
	for i := 0; i < len(matrix); i++ {
		coefficientMatrix[i] = matrix[i][:len(matrix)]
	}

	// Вычисление определителя основной матрицы (detA)
	detA, err := determinant(coefficientMatrix)
	if err != nil {
		fmt.Println("Ошибка вычисления определителя:", err)
		return
	}

	fmt.Printf("Определитель основной матрицы (detA): %.2f\n", detA)


	// Проверка условия применимости метода Крамера: определитель не должен быть равен 0
	if detA == 0 {
	if detA == 0 {
		fmt.Println("Определитель основной матрицы равен 0. Метод Крамера неприменим для нахождения единственного решения.")
		fmt.Println("Система может быть несовместной или иметь бесконечное количество решений.")
		return
	}

	startTime := time.Now() // Записываем время начала выполнения алгоритма
	solution, err := cramer(matrix)
	endTime := time.Now()                 // Записываем время окончания выполнения алгоритма
	elapsedTime := endTime.Sub(startTime) // Вычисляем разницу между временем окончания и начала

	if err != nil {
		fmt.Println("Ошибка решения методом Крамера:", err)
		return
	}

	fmt.Println("\nРешение системы уравнений методом Крамера:")
	// Вывод решения системы уравнений
	for i, x := range solution {
		fmt.Printf("x%d = %.2f\n", i+1, x)
	}

	fmt.Printf("\nВремя, затраченное на решение: %s\n", elapsedTime) // Выводим затраченное время
}


// Функция для чтения матрицы из файла
func readMatrixFromFile(filename string) ([][]float64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matrix [][]float64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		strValues := strings.Fields(line)
		var row []float64
		for _, strVal := range strValues {
			val, err := strconv.ParseFloat(strVal, 64)
			if err != nil {
				return nil, fmt.Errorf("ошибка парсинга числа: %s", strVal)
			}
			row = append(row, val)
		}
		matrix = append(matrix, row)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(matrix) == 0 {
		return nil, errors.New("файл пуст или не содержит данных")
	}
	return matrix, nil
}

// Функция для вычисления определителя матрицы рекурсивным методом (разложение по первой строке)
func determinant(matrix [][]float64) (float64, error) {
	n := len(matrix)

	if !isSquareMatrix(matrix) {
		return 0, errors.New("матрица должна быть квадратной для вычисления определителя")
	}

	// Случаи для определителей 1x1 и 2x2
	if n == 1 {
		return matrix[0][0], nil
	}
	if n == 2 {
		return matrix[0][0]*matrix[1][1] - matrix[0][1]*matrix[1][0], nil
	}

	// Рекурсивное вычисление определителя разложением по первой строке
	det := 0.0
	for j := 0; j < n; j++ {
		// Получение минора матрицы путем удаления первой строки и j-го столбца
		minorMatrix := minor(matrix, 0, j)
		// Рекурсивное вычисление определителя минора
		minorDet, err := determinant(minorMatrix)
		if err != nil {
			return 0, err
		}
		sign := 1.0
		// Определение знака в разложении (+/- в зависимости от позиции элемента)
		if j%2 == 1 {
			sign = -1.0
		}

		// Накопление суммы для общего определителя
		det += sign * matrix[0][j] * minorDet
	}
	return det, nil
}


// Функция для решения СЛАУ методом Крамера
func cramer(matrix [][]float64) ([]float64, error) {
	n := len(matrix)

	// Проверка формата матрицы для метода Крамера
	if !isAugmentedMatrixForCramer(matrix) {
		return nil, errors.New("матрица не соответствует формату для метода Крамера")
	}


	// Разделение расширенной матрицы на матрицу коэффициентов и столбец свободных членов
	coefficientMatrix := make([][]float64, n)
	constants := make([]float64, n)
	for i := 0; i < n; i++ {
		coefficientMatrix[i] = matrix[i][:n]
		constants[i] = matrix[i][n]
	}


	// Вычисление определителя основной матрицы
	detA, err := determinant(coefficientMatrix)
	if err != nil {
		return nil, err
	}

	// Проверка, что определитель не равен 0 (условие единственности решения)
	if detA == 0 {
		return nil, errors.New("определитель основной матрицы равен 0, метод Крамера неприменим для единственного решения")
	}

	solution := make([]float64, n)

	// Вычисление определителей для каждого x_i
	for i := 0; i < n; i++ {
		// Создание временной матрицы путем замены i-го столбца основной матрицы на столбец свободных членов
		tempMatrix := make([][]float64, n)
		for r := 0; r < n; r++ {
			tempMatrix[r] = make([]float64, n)
			copy(tempMatrix[r], coefficientMatrix[r])
			tempMatrix[r][i] = constants[r]
		}

		// Вычисление определителя временной матрицы (detAi)
		detAi, err := determinant(tempMatrix)
		if err != nil {
			return nil, err
		}

		/ Вычисление значения x_i = detAi / detA
		solution[i] = detAi / detA
	}
	return solution, nil
}


// Функция для получения минора матрицы путем удаления указанной строки и столбца
func minor(matrix [][]float64, row int, col int) [][]float64 {
	n := len(matrix)
	minorMatrix := make([][]float64, n-1)
	for i := 0; i < n-1; i++ {
		minorMatrix[i] = make([]float64, n-1)
	}

	currentRow := 0
	for i := 0; i < n; i++ {
		if i == row {
			continue
		}
		currentCol := 0
		for j := 0; j < n; j++ {
			if j == col {
				continue
			}
			minorMatrix[currentRow][currentCol] = matrix[i][j]
			currentCol++
		}
		currentRow++
	}
	return minorMatrix
}

// Функция для проверки, является ли матрица квадратной
func isSquareMatrix(matrix [][]float64) bool {
	if len(matrix) == 0 {
		return false
	}
	rows := len(matrix)
	cols := len(matrix[0])

	// Проверка, что все строки имеют одинаковую длину
	for _, row := range matrix {
		if len(row) != cols {
			return false
		}
	}

	// Квадратная, если количество строк равно количеству столбцов
	return rows == cols
}

// Функция для проверки, является ли матрица расширенной матрицей для метода Крамера
func isAugmentedMatrixForCramer(matrix [][]float64) bool {
	if len(matrix) == 0 {
		return false
	}
	rows := len(matrix)
	cols := 0
	if rows > 0 {
		cols = len(matrix[0])
	}
	for _, row := range matrix {
		if len(row) != cols {
			return false
		}
	}
	return rows > 0 && cols == rows+1
}
