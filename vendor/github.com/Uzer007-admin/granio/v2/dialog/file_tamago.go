//go:build tamago || noos

package dialog

import "github.com/Uzer007-admin/granio/v2"

func getFavoriteLocations() (map[string]granio.ListableURI, error) {
	return map[string]granio.ListableURI{}, nil
}
