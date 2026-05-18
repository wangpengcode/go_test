package slice

func SlicePractice() {
	array := []string{"one", "two", "three", "four", "five", "six"}
	slice := array[0:2]
	for i := range slice {
		// it will print one \n two \n
		println("slice1 ", slice[i])
	}
	println("")
	slice2 := array[2:5]
	for i := range slice {
		// "three", "four", the order come from 2-5 but not include the 5, so it include "three","four"
		println("slice2 ", slice2[i])
	}
}
