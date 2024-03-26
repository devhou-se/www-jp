function loadImageInStages(imageElement, a, b, c, d) {
    const loadNextSrc = (src, nextSrcCallback) => {
        const img = new Image();
        img.src = src;
        img.onload = () => {
            imageElement.src = src;
            if (nextSrcCallback) nextSrcCallback();
        };
    };

    loadNextSrc(a, () => loadNextSrc(b, () => loadNextSrc(c, () => loadNextSrc(d))));
}
