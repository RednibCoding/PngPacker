# PngPacker
PngPacker is a tool that can pack png files into a single file and also unpack png's from any file that contains png's

# What is PngPacker
PngPacker by Michael Binder

PngPacker is a tool that scans the given file for png signatures and if given, extracts those png files.
It also can pack a given folder of png files into a pack file. While packing, it writes a PngPacker signature
so it knows when unpacking, that it is a file packed by PngPacker. The signature is an arbitrary sequence of five bytes at the beginning see: getPngPackerSignature()
Also it packs the png file names in front of the png signatures to preserve the original file names of the png files.
So the file structure of a packed file looks like as follows:
```
 --------------------
| PngPackerSignature |  -> the signature to identify wether it is a packed file from PngPacker
| #myImage1.png#     |  -> original file name of the first image
| png signature      |  -> png signature of the first image
| png content        |  -> png content of the first image
| #myImage2.png#     |  -> original file name of the second image
| png signature      |  -> png signature of the second image
| png content        |  -> png content of the second image
| ...                |
 --------------------
```
 PngPacker can not only unpack png images from its own produced pack files but any file that contains png signatures and content.
 PngPacker can identify automatically wether it is a file packed by itself or a random other file.

# How to
## Pack png files
- drag and drop a folder containing png files onto `PngPacker`
## Unpack png files
- drag and drop a packed file onto `PngPacker` (this can be any file, because whenever `PngPacker` finds png signatures, it will try to unpack the png file)
