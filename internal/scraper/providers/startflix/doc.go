// Package startflix scrapes the StartFlix movie/TV source: catalog search on the
// WordPress front (startflix.biz), the per-title video panel it embeds
// (seasons, Dublado/Legendado episode lists, player buttons), and stream
// resolution for the player hosts that can be resolved over plain HTTP. It is a
// leaf provider package — it depends only on netx/util/models, never on the
// dispatch layers above it.
//
// There is no browser anywhere in this package: neither the site nor its panel
// is behind a bot gate.
package startflix
