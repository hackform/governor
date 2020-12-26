package rank

import (
	"net/http"
	"regexp"
	"sort"
	"strings"
	"xorkevin.dev/governor"
)

const (
	rankLengthCap = 128
)

// Tags for user rank
const (
	TagUser       = "user"
	TagUserPrefix = "usr"
	TagBanPrefix  = "ban"
	TagModPrefix  = "mod"
	TagOrgPrefix  = "org"
	TagAdmin      = "admin"
	TagSystem     = "system"
)

const (
	rankSeparator = "."
)

type (
	// Rank represents the set of all user auth tags
	Rank map[string]struct{}
)

// Len returns the size of the rank
func (r Rank) Len() int {
	return len(r)
}

// ToSlice returns an alphabetically sorted string slice of the rank
func (r Rank) ToSlice() []string {
	keys := make([]string, 0, len(r))
	for k := range r {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// String transforms the rank into an alphabetically ordered, comma delimited string
func (r Rank) String() string {
	return strings.Join(r.ToSlice(), ",")
}

// Has checks if a Rank has a tag
func (r Rank) Has(tag string) bool {
	_, ok := r[tag]
	return ok
}

// HasMod checks if a Rank has a moderator tag
func (r Rank) HasMod(tag string) bool {
	_, ok := r[TagModPrefix+rankSeparator+tag]
	return ok
}

// HasUser checks if a Rank has a user tag
func (r Rank) HasUser(tag string) bool {
	_, ok := r[TagUserPrefix+rankSeparator+tag]
	return ok
}

// HasBan checks if a Rank has a ban tag
func (r Rank) HasBan(tag string) bool {
	_, ok := r[TagBanPrefix+rankSeparator+tag]
	return ok
}

// ToModName creates a mod name from a string
func ToModName(tag string) string {
	return TagModPrefix + rankSeparator + tag
}

// ToUsrName creates a usr name from a string
func ToUsrName(tag string) string {
	return TagUserPrefix + rankSeparator + tag
}

// ToBanName creates a ban name from a string
func ToBanName(tag string) string {
	return TagBanPrefix + rankSeparator + tag
}

// AddMod adds a mod tag
func (r Rank) AddMod(tag string) Rank {
	r[ToModName(tag)] = struct{}{}
	return r
}

// AddUser adds a user tag
func (r Rank) AddUser(tag string) Rank {
	r[ToUsrName(tag)] = struct{}{}
	return r
}

// AddBan adds a ban tag
func (r Rank) AddBan(tag string) Rank {
	r[ToBanName(tag)] = struct{}{}
	return r
}

// AddOne adds a tag
func (r Rank) AddOne(tag string) Rank {
	r[tag] = struct{}{}
	return r
}

// Add adds a rank
func (r Rank) Add(other Rank) {
	for k := range other {
		r[k] = struct{}{}
	}
}

// Remove removes a rank
func (r Rank) Remove(other Rank) {
	for key := range other {
		delete(r, key)
	}
}

// Intersect returns the intersection between Ranks
func (r Rank) Intersect(other Rank) Rank {
	intersect := Rank{}
	for k := range other {
		if _, ok := r[k]; ok {
			intersect[k] = struct{}{}
		}
	}
	return intersect
}

var (
	rankRegexMod = regexp.MustCompile(`^mod.[A-Za-z0-9.-_]+$`)
	rankRegexUsr = regexp.MustCompile(`^usr.[A-Za-z0-9.-_]+$`)
	rankRegexBan = regexp.MustCompile(`^ban.[A-Za-z0-9.-_]+$`)
	rankRegexOrg = regexp.MustCompile(`^org.[A-Za-z0-9.-_]+$`)
)

// FromSlice creates a new Rank from a list of strings
func FromSlice(rankSlice []string) Rank {
	if len(rankSlice) < 1 {
		return Rank{}
	}
	r := make(Rank, len(rankSlice))
	for _, i := range rankSlice {
		r[i] = struct{}{}
	}
	return r
}

// FromStringUser creates a new User Rank from a string
func FromStringUser(rankString string) (Rank, error) {
	if len(rankString) < 1 {
		return Rank{}, nil
	}
	rankSlice := strings.Split(rankString, ",")
	for _, i := range rankSlice {
		if len(i) > rankLengthCap || !rankRegexMod.MatchString(i) && !rankRegexUsr.MatchString(i) && !rankRegexBan.MatchString(i) && i != TagUser && i != TagAdmin && i != TagSystem {
			return Rank{}, governor.NewError("Illegal rank string", http.StatusBadRequest, nil)
		}
	}
	return FromSlice(rankSlice), nil
}

func SplitTag(key string) (string, string, error) {
	k := strings.SplitN(key, rankSeparator, 2)
	if len(k) < 2 {
		return "", "", governor.NewError("Illegal rank string", http.StatusBadRequest, nil)
	}
	switch k[0] {
	case TagModPrefix, TagUserPrefix, TagBanPrefix:
	default:
		return "", "", governor.NewError("Illegal rank string", http.StatusBadRequest, nil)
	}
	return k[0], k[1], nil
}

// ToOrgName creates a new org name from a string
func ToOrgName(name string) string {
	return TagOrgPrefix + rankSeparator + name
}

// IsValidOrgName validates an orgname
func IsValidOrgName(orgname string) bool {
	return rankRegexOrg.MatchString(orgname)
}

// BaseUser creates a new user rank
func BaseUser() Rank {
	return Rank{
		TagUser: struct{}{},
	}
}

// Admin creates a new Administrator rank
func Admin() Rank {
	b := BaseUser()
	b[TagAdmin] = struct{}{}
	return b
}

// System creates a new System rank
func System() Rank {
	return Rank{
		TagSystem: struct{}{},
	}
}
