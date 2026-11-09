// Package goannoy defines common values for the metric-specific Annoy packages.
package goannoy

// ItemID identifies an item in the standard Spotify Annoy int32 format.
type ItemID int32

// Result is an item and its distance (or similarity score for dot product).
type Result struct {
	ID       ItemID
	Distance float32
}

// SearchK is the candidate budget used during approximate tree traversal.
type SearchK int

// AutoSearchK uses len(destination) multiplied by the number of trees.
const AutoSearchK SearchK = -1
