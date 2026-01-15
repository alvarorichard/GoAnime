pkgname=goanime
pkgver=1.4
pkgrel=1
pkgdesc="A command-line tool to browse, play, and download anime."
arch=('x86_64')
url="https://github.com/alvarorichard/GoAnime"
license=('MIT')
depends=('mpv')
source=("goanime-linux::https://github.com/alvarorichard/GoAnime/releases/download/v$pkgver/goanime-linux")
sha256sums=('SKIP')  # Ignora validação de checksum


package() {
  install -Dm755 "$srcdir/goanime-linux" "$pkgdir/usr/bin/goanime"
}
