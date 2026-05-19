package _interface

import "math"

type Geometry interface {
	Area() float64
	Perimeter() float64
}

type Rectangle struct {
	Width  float64
	Height float64
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

type Circle struct {
	Radius float64
}

func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

func GeometryTest() {
	rectangle := Rectangle{Width: 10, Height: 5}
	circle := Circle{Radius: 5}

	println("the circle have radius", 5, " and it's area is ", circle.Area())
	println("the circle have radius", 5, " and it's perimeter is ", circle.Perimeter())

	println("the rectangle have width,", 10, " and it's height is ", 5, " and we can find is area is ", rectangle.Area())
	println("the rectangle have width,", 10, " and it's height is ", 5, " and we can find is perimeter is ", rectangle.Perimeter())
}
