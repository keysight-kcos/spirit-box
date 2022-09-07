A spirit-box binary should be placed in `<package-dir>/usr/bin/` or `<package-dir>/root/usr/bin/` first.

`dpkg-deb --build ./<package-dir>`

`dpkg -i ./<package-name>.deb`
