/*
Copyright 2017 Nirmal Kumar

This file is part of LibreRead.

LibreRead is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

LibreRead is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with LibreRead.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Router
	r.GET("/", GetQuote)

	// Listen and serve on 0.0.0.0:8080
	r.Run(":3000")
}

func GetQuote(c *gin.Context) {
	c.JSON(200, gin.H{
		"quote":       "Never doubt that a small group of thoughtful, committed, citizens can change the world. Indeed, it is the only thing that ever has.",
		"author":      "Margaret Mead",
		"authorURL":   "https://www.goodreads.com/author/show/61107.Margaret_Mead",
		"image":       "https://images.gr-assets.com/authors/1198589352p2/61107.jpg",
		"fromBook":    "",
		"fromBookURL": "",
	})
}
