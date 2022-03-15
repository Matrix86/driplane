package apt

import (
	"strings"
	"testing"
)

func TestParseRelease(t *testing.T) {
	txt := `Origin: Debian
Label: Debian
Suite: stable
Version: 11.2
Codename: bullseye
Changelogs: https://metadata.ftp-master.debian.org/changelogs/@CHANGEPATH@_changelog
Date: Sat, 18 Dec 2021 10:38:58 UTC
Acquire-By-Hash: yes
No-Support-for-architecture-all: PackagePaths
Architectures: all amd64 arm64 armel armhf i386 mips64el mipsel ppc64el s390x
Components: main contrib non-free
Description: Debian 11.2 Released 18 December 2021
MD5Sum:
 7fdf4db15250af5368cc52a91e8edbce   738242 contrib/Contents-all
 cbd7bc4d3eb517ac2b22f929dfc07b47    57319 contrib/Contents-all.gz
 37d6231ff08b9f383fba5134e90c1246   786460 contrib/Contents-amd64
 f862fd63c5e4927f91c1cc77a27c89eb    54567 contrib/Contents-amd64.gz
 098f43776cd6b43334b2c1eb867a4d64   370054 contrib/Contents-arm64
 014eca050cdd0b30df0d5a72df217c5b    29661 contrib/Contents-arm64.gz
 b6d2673f17fbdb3a5ce92404a62c2d7e   359292 contrib/Contents-armel
 d02d94be587d56a1246b407669d2a24c    28039 contrib/Contents-armel.gz
SHA256:
 3957f28db16e3f28c7b34ae84f1c929c567de6970f3f1b95dac9b498dd80fe63   738242 contrib/Contents-all
 3e9a121d599b56c08bc8f144e4830807c77c29d7114316d6984ba54695d3db7b    57319 contrib/Contents-all.gz
 425a90016c3b1b64fd560b0f4524d53d5b3ee3aa0835859408e8413aa1145dc9   786460 contrib/Contents-amd64
 1c336a418784bb0eb78318bab337b4df34b34e56683e3a7887f319a2a8985c6b    54567 contrib/Contents-amd64.gz
 d1301db9f59f4baf78398a8e123e76088b9963b612274e030e8c1d52720e0151   370054 contrib/Contents-arm64
 1a01ec345569da804ff98ed259b2135845130a30e434909c4e69c88bf9cb8d9a    29661 contrib/Contents-arm64.gz
 b4985377d670dbc4ab9bf0f7fb15d11b100c442050dee7c1e9203d3f0cfd3f37   359292 contrib/Contents-armel
 f134666bc09535cbc917f63022ea31613da15ec3c0ce1c664981ace325acdd6a    28039 contrib/Contents-armel.gz`

	r := strings.NewReader(txt)
	release, err := ParseRelease(r)
	if err != nil {
		t.Errorf("error received: %s", err)
	}
	if len(release.MD5Sum) != 8 {
		t.Errorf("wrong num of MD5Sum")
	}
	if len(release.SHA256) != 8 {
		t.Errorf("wrong num of SHA256")
	}
	if len(release.PackagePaths) != 8 {
		t.Errorf("wrong num of paths")
	}
}
