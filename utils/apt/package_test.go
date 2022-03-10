package apt

import (
	"strings"
	"testing"
)

func TestParsePackage(t *testing.T) {
	txt := `BinaryPackage: me.absidue.ignoexplore
architecture: iphoneos-arm
Version: 0.0.6
Section: Tweaks
Maintainer: absidue
Installed-Size: 88
Depends: firmware (>= 13.0), org.coolstar.libhooker | mobilesubstrate
Filename: ./pool/me.absidue.ignoexplore_0.0.6_iphoneos-arm.deb
Size: 3750
MD5sum: 99b33f7a6b0ff67be3603d1a5cdf57de
SHA1: d0656d068121bd82c81d425ef4af9de9222ebd21
SHA256: d84f3843586229f3f5682a301e23e25a34d9b86e0296ebf99d6ffdddc0c3101f
SHA512: 97cfee4cdd3c7fec1f31a23110bc5c84032f0b6126f353b82ad7da9d97c85c55a588955b38b54b7f8c83f27cd2bf9ede1276161a1c1554a11cfbace7adba9dc5
Homepage: https://github.com/absidue/IGNoExplore/
Description: Hides the explore feed in Instagram and enters search field straight away
 Hides the explore feed on the Instagram search page and focuses the search field immediately.
 .
 Works with the latest Instagram version at the time of release, please file an issue on the GitHub page if it's broken for you.
 .
 No options to configure.
Tag: compatible_min::ios13.0
Name: IGNoExplore
Author: absidue
Icon: https://apt.absidue.me/assets/icons/me.absidue.ignoexplore.png

BinaryPackage: me.absidue.musicrecognitionmoreapps
architecture: iphoneos-arm
Version: 1.0.1
Section: Tweaks
Maintainer: absidue
Installed-Size: 168
Depends: firmware (>= 14.0), mobilesubstrate (>= 0.9.5000)
Filename: ./pool/me.absidue.musicrecognitionmoreapps_1.0.1_iphoneos-arm.deb
Size: 8006
MD5sum: c0b64cea641a9da67ba7de1513d844eb
SHA1: 661811ebc49314e88b414971be3aa9c47d71603d
SHA256: c4a6d899f8a4cb2877bd5a8fa4a8130790dc65c805ff4cbbf9a34a24c1ca628f
SHA512: 6c31b0456cf2e923cd408837a1db61daac927ca5e3387201893de71a7fa5e1ede29a667d9a16fc4b29d28418a4a8748b150511746820d598c28b16c6a74a0134
Homepage: https://github.com/absidue/MusicRecognitionMoreApps/
Description: Open Music Recognition results in more apps
 Open Music Recognition results in more apps.
 Automatically detects which apps you have installed and shows the relevant apps.
 .
 Currently supported apps:
 - YouTube
 - YouTube Music
 - Spotify
 - Deezer
 .
 If your music app is not supported yet, please open an issue or pull request on the GitHub page.
 .
 No options to configure.
Tag: compatible_min::ios14.0
Name: MusicRecognitionMoreApps
Author: absidue

BinaryPackage: me.absidue.musicrecognitionmoreapps
architecture: iphoneos-arm
Version: 1.0
Section: Tweaks
Maintainer: absidue
Installed-Size: 168
Depends: firmware (>= 14.0), mobilesubstrate (>= 0.9.5000)
Filename: ./pool/me.absidue.musicrecognitionmoreapps_1.0_iphoneos-arm.deb
Size: 7088
MD5sum: d7e8d18c16fb947d923969af9f82ba4d
SHA1: 5762101b97655fc7a2313cc75d3f5ebff15c03c3
SHA256: fefc36170b48e391c96a1a8c3aac78d1ab3f95b9eb461408c22be53e039eb570
SHA512: 8d8784919fadf14032889abb6c5e3c933fd561d4da5d6f92c6c24f9ba774f15c0e539df32d0d01a6644c37984b54218a6af05b12361825587c0047ba0df9e12f
Homepage: https://github.com/absidue/MusicRecognitionMoreApps/
Description: Open Music Recognition results in more apps
 Open Music Recognition results in more apps.
 Automatically detects which apps you have installed and shows the relevant apps.
 .
 Currently supported apps:
 - YouTube
 - YouTube Music
 - Spotify
 - Deezer
 .
 If your music app is not supported yet, please open an issue or pull request on the GitHub page.
 .
 No options to configure.
Tag: compatible_min::ios14.0
Name: MusicRecognitionMoreApps
Author: absidue`

	r := strings.NewReader(txt)
	p, err := ParsePackageIndex(r, "")
	if err != nil {
		t.Errorf("error received: %s", err)
	}
	if len(p.Binaries) != 3 {
		t.Errorf("the index should contains 3 packages")
	}
}
