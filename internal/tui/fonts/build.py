#!/usr/bin/env python3
"""Builds GoAnimeGlyphs.otf, the font GoAnime installs on Linux when the
machine has no Japanese font (see internal/tui/japanese_font_linux.go).

It is Noto Sans Mono CJK JP cut down to the characters in glyphs.txt and
renamed, as the SIL Open Font License asks of modified versions, so it never
stands in for the real Noto font. TestGlyphFontCoversEveryGloss fails when a
new Japanese string is not in the font; add its characters to glyphs.txt and
run, from this directory:

    pip install fonttools
    python3 build.py NotoSansMonoCJK-VF.ttc

The source is Noto Sans Mono CJK JP 2.004, static (NotoSansMonoCJKjp-Regular.otf)
or variable (NotoSansMonoCJK-VF.ttc, whose first font is JP), from
https://github.com/notofonts/noto-cjk or a distribution package.
"""

import sys

from fontTools import subset
from fontTools.ttLib import TTFont
from fontTools.varLib import instancer

FAMILY = "GoAnime Glyphs"
POSTSCRIPT = "GoAnimeGlyphs-Regular"


def main(source):
    with open("glyphs.txt", encoding="utf-8") as f:
        text = f.read().strip()

    font = TTFont(source, fontNumber=0, recalcTimestamp=False)
    if font["name"].getDebugName(1) != "Noto Sans Mono CJK JP":
        sys.exit(f"{source}: the first font is not Noto Sans Mono CJK JP")
    version = (font["name"].getDebugName(5) or "").split(";")[0]
    options = subset.Options()
    options.name_IDs = [0, 13, 14]  # copyright and license
    options.name_languages = [0x409]
    options.layout_features = []
    options.notdef_outline = True
    subsetter = subset.Subsetter(options)
    subsetter.populate(text=text)
    subsetter.subset(font)
    if "fvar" in font:
        # Regular only; the outlines stay CFF2, which FreeType reads since 2.8.
        font = instancer.instantiateVariableFont(font, {"wght": 400})

    missing = [c for c in text if ord(c) not in font.getBestCmap()]
    if missing:
        sys.exit(f"{source} has no glyph for {''.join(missing)}")

    name = font["name"]
    for platform, encoding, language in ((3, 1, 0x409), (1, 0, 0)):
        name.setName(FAMILY, 1, platform, encoding, language)
        name.setName("Regular", 2, platform, encoding, language)
        name.setName(f"{FAMILY}: {version}", 3, platform, encoding, language)
        name.setName(FAMILY, 4, platform, encoding, language)
        name.setName(version, 5, platform, encoding, language)
        name.setName(POSTSCRIPT, 6, platform, encoding, language)
    font["OS/2"].achVendID = "NONE"
    font.save("GoAnimeGlyphs.otf")
    print(f"GoAnimeGlyphs.otf: {len(text)} characters")


if __name__ == "__main__":
    if len(sys.argv) != 2:
        sys.exit(__doc__)
    main(sys.argv[1])
