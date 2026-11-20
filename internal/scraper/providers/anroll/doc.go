// Package anroll implements the plain-HTTP scraper for anroll.io and its
// anidrive.click player.
//
// AnimesROLL publishes subtitled and dubbed releases as separate title pages.
// Search merges matching pairs and carries the dubbed page in an opaque URL
// fragment so the API layer can ask for the desired audio only when both exist.
//
// Episode pages embed an anidrive.click token page. Its inline player config is
// base64-split and XOR-obfuscated; this package decodes that data without
// executing JavaScript and selects the best HLS or MP4 source.
package anroll
