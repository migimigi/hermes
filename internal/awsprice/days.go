package awsprice

func GetDays(i int) int {
	return Days[i%12]
}

var Days = map[int]int{
	0:  31, //12
	1:  31,
	2:  28,
	3:  31,
	4:  30,
	5:  31,
	6:  30,
	7:  31,
	8:  31,
	9:  30,
	10: 31,
	11: 30,
}