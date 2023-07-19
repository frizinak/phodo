# Phổ Đỏ

An image processing library and pipeline scripting language.

## Goal

Simple yet performant library to create an equally simple RAW/JPEG editor to 
integrate into my photography workflow.

## To do

- [ ] Decently sized stdlib
    - [X] white-balance-spot(x y radius)
    - [ ] visualize-clipping()
    - [ ] spot removal / clone tool
    - [ ] reduce-noise()
    - [ ] exif add/set/delete
    - [ ] glob() / find() / exec(`find -iname '*.png'`)
    - [ ] dcraw
    - [ ] ...
- [ ] More scripting options
    - [X] Include files
    - [X] math
    - [X] variables
    - [ ] defining composite elements
    - [ ] ...
- [X] EXIF input
- [ ] EXIF output
- [ ] Documentation

## Scripting example

```
.a-clut(clut(cache(load-file("cluts/a hald clut.png"))))

# ${my-includes-dir}/include.pho

.colors(
    rgb-multiply(0.90 0.85 1.3)
    saturation(0.80)
    gamma(1.3)
    contrast(0.25)
    black(0.10)
)

.main(
    // Try to load an .i48 internal cache file.
    // Load the actual tif/jpeg/... if it doesn't exist,
    // correct the orientation and store in the .i48 cache file.
    or(
        load-file("${file}.i48")
        (
            load-file("${file}")
            orientation()
            save-file("${file}.i48")
        )
    )

    crop(0 1200 -1 3000)

    // Save our original, slightly cropped image in memory variable 'original'
    save(original)

    // Crop out the eye of the lizard and store in memory variable 'eye'
    // after adjusting some of the colors.
    crop(1770 570 310 310)
    .colors()
    gamma(1.7)
    black(0.2)
    save(eye)

    // Start working from 'original' again.
    load(original)
    // Resize to desired output size and store in 'img'.
    resize-fit(2160 2160)
    save(img)

    // Create a histgram from our image before we adjust the colors.
    histogram(rgb 300 150 5 2)
    save(histogram)

    // White balance and increase contrast etc.
    // in the same way we did it for the 1:1 pixel ratio sub image 'eye'
    load(img)
    .colors()

    // Example of a tee element, not strictly necessary but is an alternative
    // to save and load calls in some use-cases.
    // None of the changes within a tee'd pipeline affect the image of the
    // pipeline surrounding it.
    // Well logically, technically no operation requires copying the pixel
    // array of our image, so any change withing tee could be reflected
    // on the tee input image. To ensure you do not change the input image,
    // a call to 'copy()' is necessary.
    tee(
        copy()
        resize-clip(310 0)
        crop(0 0 310 200)
        saturation(0)
        save-file("data/thumb.jpg" 80)
        save(thumb)
    )
    compose(
        pos-alpha(50 50 (
                load(histogram)
            )
        )
    )
    .a-clut()
    compose(
        pos(50 250 load(thumb))
        pos-alpha(50 475 histogram(rgb 300 150 5 2))
        pos(50 675 load(eye))
    )
    save-file("data/result.jpg")
)
```

`phodo do script.pho file="input.tif"`

![example result](https://raw.githubusercontent.com/frizinak/phodo/dev/.github/main.jpg)
