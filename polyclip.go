package polyclip

import "reflect"

type Point struct {
	X float64
	Y float64
}

type Polygon struct {
	Points []Point
}

type Node struct {
	Poly      Polygon
	Index     int
	Point     Point
	Intersect bool
	Dist      float64
	Next      *Node
	Prev      *Node
	IsEntry   bool
	Friend    *Node
	Processed bool
}

type Intersection struct {
	Cross  float64
	AlongA float64
	AlongB float64
	Point  Point
}

type Result struct {
	Result  Polygon
	RootOne *Node
	RootTwo *Node
}

func UpgradePolygon(polygon Polygon) *Node {

	// converts a list of points into a double linked list
	var root *Node
	var prev *Node

	for i := 0; i < len(polygon.Points); i++ {

		node := &Node{

			Poly:      polygon, // extra data to help the UI -- unnecessary for algorithm to work:
			Index:     i,       // extra data to help the UI -- unnecessary for algorithm to work:
			Point:     polygon.Points[i],
			Intersect: false,
			Next:      nil,
			Prev:      nil,
		}

		if root == nil { // root just points to itself:

			node.Next = node
			node.Prev = node
			root = node

		} else {

			prev = root.Prev // change this:
			prev.Next = node //    ...-- (prev) <--------------> (root) --...
			node.Prev = prev // to this:
			node.Next = root //    ...-- (prev) <--> (node) <--> (root) --...
			root.Prev = node

		}

	}

	return root

}

func LinesIntersect(aZero Point, aOne Point, bZero Point, bOne Point) Intersection {

	adX := aOne.X - aZero.X
	adY := aOne.Y - aZero.Y
	bdX := bOne.X - bZero.X
	bdY := bOne.Y - bZero.Y

	axb := adX*bdY - adY*bdX

	intersect := Intersection{}

	if axb == 0 {

		return intersect

	}

	dx := aZero.X - bZero.X
	dy := aZero.Y - bZero.Y

	intersect.AlongA = (bdX*dy - bdY*dx) / axb
	intersect.AlongB = (adX*dy - adY*dx) / axb

	intersect.Point = Point{
		X: aZero.X + intersect.AlongA*adX,
		Y: aZero.Y + intersect.AlongA*adY,
	}

	return intersect

}

func NextNonIntersection(node *Node) *Node {

	for node.Intersect {

		node = node.Next

	}

	return node

}

func PointInPolygon(point Point, root *Node) bool {

	isOdd := false
	x := point.X
	y := point.Y
	here := root

	var (
		next *Node
		hx   float64
		hy   float64
		nx   float64
		ny   float64
		pip  bool
	)

	for reflect.DeepEqual(here, root) == false {

		next = here.Next
		hx = here.Point.X
		hy = here.Point.Y
		nx = next.Point.X
		ny = next.Point.Y

		pip = (((hy < y && ny >= y) || (hy >= y && ny < y)) &&
			(hx <= x || nx <= x) &&
			(hx+(y-hy)/(ny-hy)*(nx-hx) < x))

		if pip {

			isOdd = !isOdd

		}

		here = next

	}

	return isOdd

}

func CalculateEntryExit(root *Node, isEntry bool) {

	here := root

	for reflect.DeepEqual(here, root) == false {

		if here.Intersect {

			here.IsEntry = isEntry
			isEntry = !isEntry

		}

		here = here.Next

	}

}

func PolygonClip(polyOne Polygon, polyTwo Polygon, intoRed bool, intoBlue bool) Result {

	rootOne := UpgradePolygon(polyOne)
	rootTwo := UpgradePolygon(polyTwo)

	// do this before inserting intersections, for simplicity
	isOneInTwo := PointInPolygon(rootOne.Point, rootTwo)
	isTwoInOne := PointInPolygon(rootTwo.Point, rootOne)

	hereOne := rootOne
	hereTwo := rootTwo

	for reflect.DeepEqual(hereOne, rootOne) {

		for reflect.DeepEqual(hereTwo, rootTwo) {

			var next1 = NextNonIntersection(hereOne)
			var next2 = NextNonIntersection(hereTwo)

			var intersect = LinesIntersect(
				hereOne.Point, next1.Point,
				hereTwo.Point, next2.Point,
			)

			if intersect.AlongA > 0 && intersect.AlongA < 1 &&
				intersect.AlongB > 0 && intersect.AlongB < 1 {

				nodeOne := &Node{
					Point:     intersect.Point,
					Intersect: true,
					Next:      nil,
					Prev:      nil,
					Dist:      intersect.AlongA,
					Friend:    nil,
				}

				nodeTwo := &Node{
					Point:     intersect.Point,
					Intersect: true,
					Next:      nil,
					Prev:      nil,
					Dist:      intersect.AlongB,
					Friend:    nil,
				}

				// point the nodes at each other
				nodeOne.Friend = nodeTwo
				nodeTwo.Friend = nodeOne

				var inext *Node
				var iprev *Node

				// find insertion between hereOne and next1, based on dist
				inext = hereOne.Next
				for inext != next1 && inext.Dist < nodeOne.Dist {
					inext = inext.Next
				}

				iprev = inext.Prev

				// insert nodeOne between iprev and inext
				inext.Prev = nodeOne
				nodeOne.Next = inext
				nodeOne.Prev = iprev
				iprev.Next = nodeOne

				// find insertion between hereTwo and next2, based on dist
				inext = hereTwo.Next

				for inext != next2 && inext.Dist < nodeTwo.Dist {
					inext = inext.Next
				}

				iprev = inext.Prev

				// insert nodeTwo between iprev and inext
				inext.Prev = nodeTwo
				nodeTwo.Next = inext
				nodeTwo.Prev = iprev
				iprev.Next = nodeTwo

			}

			hereTwo = NextNonIntersection(hereTwo)

		}

		hereOne = NextNonIntersection(hereOne)

	}

	CalculateEntryExit(rootOne, !isOneInTwo)
	CalculateEntryExit(rootTwo, !isTwoInOne)

	var result Polygon
	isect := rootOne
	into := [2]bool{intoBlue, intoRed}

	for true {

		for isect != rootOne {

			if isect.Intersect && !isect.Processed {
				break
			}

			isect = isect.Next

		}

		if isect == rootOne {

			break

		}

		curpoly := 0
		var clipped Polygon

		here := isect

		for !here.Processed {

			// mark intersection as processed
			here.Processed = true
			here.Friend.Processed = true

			var moveForward = here.IsEntry == into[curpoly]

			for !here.Intersect {

				clipped.Points = append(clipped.Points, here.Point)

				if moveForward {

					here = here.Next

				} else {

					here = here.Prev

				}

				// we've hit the next intersection so switch polygons
				here = here.Friend
				curpoly = 1 - curpoly

			}

			result.Points = append(result.Points, clipped.Points...)

		}

		if len(result.Points) <= 0 {

			if isOneInTwo == intoBlue {

				result.Points = append(result.Points, polyOne.Points...)
			}

			if isTwoInOne == intoRed {

				result.Points = append(result.Points, polyTwo.Points...)

			}

		}

	}
	
	return Result{
			RootOne: rootOne, // used for UI
			RootTwo: rootTwo, // used for UI
			Result:  result,  // this is all you really need
		}

}

