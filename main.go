package main

import (
	"fmt"
	"math"
)

func cal(x int, y int, d int) int {
	inc, dec := 0, 0
	if x > y {
		inc = y + d - x
		dec = x - y
	} else {
		inc = y - x
		dec = x + d - y
	}
	return min(inc, dec)
}

func min(x, y int) int {
	if x > y {
		return y
	} else {
		return x
	}
}

func main() {
	T := 0
	fmt.Scanf("%d", &T)
	for t := 0; t < T; t++ {
		n, d := 0, 0
		fmt.Scanf("%d%d", &n, &d)
		arr := make([]int, n)
		for i := range arr {
			fmt.Scanf("%d", &arr[i])
		}
		dp := make([][][]int, n)
		for i := 0; i < n; i++ {
			dp[i] = make([][]int, n)
			for j := range dp[i] {
				dp[i][j] = make([]int, 2)
				dp[i][j][0] = math.MaxInt
				dp[i][j][1] = math.MaxInt
			}
			dp[i][i][0] = 0
			dp[i][i][1] = 0
		}
		for l := 1; l < n; l++ {
			for i := 0; i+l < n; i++ {
				j := i + l
				dp[i][j][0] = min(dp[i+1][j][0]+cal(arr[i+1], arr[i], d), dp[i+1][j][1]+cal(arr[j], arr[i], d))
				dp[i][j][1] = min(dp[i][j-1][0]+cal(arr[i], arr[j], d), dp[i][j-1][1]+cal(arr[j-1], arr[j], d))
			}
		}
		ans := min(dp[0][n-1][0]+cal(arr[0], 0, d), dp[0][n-1][1]+cal(arr[n-1], 0, d))
		fmt.Printf("Case #%d: %d\n", t+1, ans)
	}
}
