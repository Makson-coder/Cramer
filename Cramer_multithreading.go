package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Функция для вычисления определителя матрицы
func determinant(matrix [][]float64) float64 {
	n := len(matrix)
	//случай: матрица 1x1. Определитель - единственный элемент.
	if n == 1 {
		return matrix[0][0]
	}

	//случай: матрица 2x2. Определитель вычисляется по формуле ad - bc.
	if n == 2 {
		return matrix[0][0]*matrix[1][1] - matrix[0][1]*matrix[1][0]
	}

	// Инициализация определителя
	// Рекурсивный шаг: разложение по первой строке.
	det := 0.0
	for j := 0; j < n; j++ {
		subMatrix := make([][]float64, n-1) // Создание подматрицы для минора
		for i := 1; i < n; i++ {
			subMatrix[i-1] = make([]float64, n-1)
			col := 0
			for k := 0; k < n; k++ {
				if k != j { // Исключаем j-й столбец для подматрицы
					subMatrix[i-1][col] = matrix[i][k]
					col++
				}
			}
		}

		// Рекурсивное вычисление определителя подматрицы и добавление к общему определителю.
		// Используется формула разложения с чередованием знаков (-1)^j.
		det += math.Pow(-1, float64(j)) * matrix[0][j] * determinant(subMatrix)
	}
	return det
}

// Функция для замены столбца в матрице
func replaceColumn(matrix [][]float64, column []float64, colIndex int) [][]float64 {
	n := len(matrix)
	newMatrix := make([][]float64, n) // Создание новой матрицы
	for i := 0; i < n; i++ {
		newMatrix[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			newMatrix[i][j] = matrix[i][j] // Копирование элементов из исходной матрицы
		}
	}

	// Замена столбца colIndex на вектор column.
	for i := 0; i < n; i++ {
		newMatrix[i][colIndex] = column[i]
	}
	return newMatrix
}

// Функция для решения СЛАУ методом Крамера с параллельными вычислениями
func cramerParallel(matrix [][]float64, vector []float64) []float64 {
	n := len(matrix)
	detMain := determinant(matrix) // Вычисление главного определителя матрицы коэффициентов.

	// Проверка на вырожденность матрицы (определитель равен 0).
	// Если определитель равен 0, система либо не имеет решений, либо имеет бесконечно много решений.
	if detMain == 0 {
		return nil // Система не имеет единственного решения или несовместна
	}

	solutions := make([]float64, n)
	var wg sync.WaitGroup         // WaitGroup для ожидания завершения всех горутин.
	results := make(chan struct { // Канал для передачи результатов из горутин (в данном случае передаем пустую структуру, т.к. порядок важен).
		index    int     // Индекс переменной, для которой вычислено решение.
		solution float64 // Значение решения для переменной.
	}, n)

	// Параллельное вычисление определителей для каждого столбца.
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(index int) { // Запуск горутины для вычисления решения для i-й переменной.
			defer wg.Done()                                        // Уменьшаем счетчик WaitGroup при завершении горутины.
			modifiedMatrix := replaceColumn(matrix, vector, index) // Создание матрицы путем замены i-го столбца на вектор правых частей.
			detModified := determinant(modifiedMatrix)             // Вычисление определителя модифицированной матрицы.
			solutions[index] = detModified / detMain               // Вычисление значения i-й переменной по формуле Крамера.
			results <- struct {                                    // Отправка результата в канал.
				index    int
				solution float64
			}{index: index, solution: detModified / detMain}
		}(i)
	}

	wg.Wait()      // Ожидание завершения всех горутин.
	close(results) // Закрытие канала results после завершения всех горутин.

	return solutions
}

func main() {
	var matrix [][]float64
	var vector []float64
	var err error

	// Чтение матрицы и вектора из файла "matrix.txt".
	matrix, vector, err = readMatrixAndVectorFromFile("matrix.txt")
	if err != nil {
		fmt.Println("Ошибка чтения данных из файла:", err)
		return
	}

	startTime := time.Now()                    // Запись времени начала выполнения алгоритма.
	solution := cramerParallel(matrix, vector) // Решение СЛАУ методом Крамера.
	endTime := time.Now()                      // Запись времени окончания выполнения алгоритма.

	// Проверка, было ли найдено решение (определитель главной матрицы не равен 0).
	if solution != nil {
		fmt.Println("Решение системы:")
		for i, sol := range solution {
			fmt.Printf("x%d = %f\n", i+1, sol)
		}
	} else {
		fmt.Println("Система не имеет единственного решения или несовместна.")
	}

	fmt.Println("\nВремя выполнения алгоритма:", endTime.Sub(startTime))
}

// Функция для чтения матрицы и вектора из файла matrix.txt, размерность определяется автоматически
func readMatrixAndVectorFromFile(filename string) ([][]float64, []float64, error) {
	file, err := os.Open(filename) // Открытие файла для чтения.
	if err != nil {
		return nil, nil, fmt.Errorf("не удалось открыть файл: %w", err)
	}
	defer file.Close() // Закрытие файла после завершения работы функции.

	scanner := bufio.NewScanner(file)

	matrixRowsStr := []string{} // Слайс для временного хранения строк матрицы (в виде строк).
	for scanner.Scan() {        // Чтение файла построчно.
		line := scanner.Text()
		if line == "" { // Пропускаем пустые строки, если есть
			continue
		}
		if !strings.Contains(line, " ") { // Предполагаем, что последняя строка - это вектор, если в ней нет пробелов
			vectorStr := line
			vectorValuesStr := strings.Split(vectorStr, " ")

			n := len(matrixRowsStr) // Размер матрицы определяется количеством строк матрицы
			if n == 0 {
				return nil, nil, fmt.Errorf("не найдено строк матрицы перед вектором")
			}

			vector = make([]float64, n) // Инициализация матрицы коэффициентов.
			if len(vectorValuesStr) != n {
				return nil, nil, fmt.Errorf("неверное количество элементов в векторе, ожидалось %d, получено %d", n, len(vectorValuesStr))
			}
			for i := 0; i < n; i++ {
				val, err := strconv.ParseFloat(vectorValuesStr[i], 64)
				if err != nil {
					return nil, nil, fmt.Errorf("ошибка парсинга элемента вектора [%d]: %w", i+1, err)
				}
				vector[i] = val
			}

			matrix = make([][]float64, n)
			for i := 0; i < n; i++ {
				matrix[i] = make([]float64, n)
				rowValuesStr := strings.Split(matrixRowsStr[i], " ")
				if len(rowValuesStr) != n {
					return nil, nil, fmt.Errorf("неверное количество элементов в строке матрицы %d, ожидалось %d, получено %d", i+1, n, len(rowValuesStr))
				}
				for j := 0; j < n; j++ {
					val, err := strconv.ParseFloat(rowValuesStr[j], 64)
					if err != nil {
						return nil, nil, fmt.Errorf("ошибка парсинга элемента матрицы [%d][%d]: %w", i+1, j+1, err)
					}
					matrix[i][j] = val
				}
			}
			return matrix, vector, nil // Возвращаем матрицу и вектор после обработки всего файла

		} else {
			matrixRowsStr = append(matrixRowsStr, line) // Сохраняем строки, которые выглядят как строки матрицы
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("ошибка сканирования файла: %w", err)
	}

	return nil, nil, fmt.Errorf("не найден вектор в файле или файл имеет неверный формат") // Если вектор так и не был найден
}
