//go:generate granio bundle --package=faces --prefix=card -o bundled.go .

package faces

import "github.com/Uzer007-admin/granio/v2"

// ForCard returns the face resource for the specified card value and suit.
func ForCard(card, suit int) granio.Resource {
	return faceResources[card-1+(suit*13)]
}

// ForBack returns a face resource for the back of a card
func ForBack() granio.Resource {
	return cardBackSvg
}

// ForSpace returns a special resource to use when a vacant spot should be indicated
func ForSpace() granio.Resource {
	return cardSpaceSvg
}

var faceResources = [52]granio.Resource{
	cardACSvg,
	card2CSvg,
	card3CSvg,
	card4CSvg,
	card5CSvg,
	card6CSvg,
	card7CSvg,
	card8CSvg,
	card9CSvg,
	card10CSvg,
	cardJCSvg,
	cardQCSvg,
	cardKCSvg,

	cardADSvg,
	card2DSvg,
	card3DSvg,
	card4DSvg,
	card5DSvg,
	card6DSvg,
	card7DSvg,
	card8DSvg,
	card9DSvg,
	card10DSvg,
	cardJDSvg,
	cardQDSvg,
	cardKDSvg,

	cardAHSvg,
	card2HSvg,
	card3HSvg,
	card4HSvg,
	card5HSvg,
	card6HSvg,
	card7HSvg,
	card8HSvg,
	card9HSvg,
	card10HSvg,
	cardJHSvg,
	cardQHSvg,
	cardKHSvg,

	cardASSvg,
	card2SSvg,
	card3SSvg,
	card4SSvg,
	card5SSvg,
	card6SSvg,
	card7SSvg,
	card8SSvg,
	card9SSvg,
	card10SSvg,
	cardJSSvg,
	cardQSSvg,
	cardKSSvg,
}
