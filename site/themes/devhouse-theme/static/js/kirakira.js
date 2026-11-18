// キラキラ: sparkles trail the mouse, and the tab title scrolls like a marquee.
(function () {
    var calm = window.matchMedia && matchMedia('(prefers-reduced-motion: reduce)').matches;

    var title = document.title + '　☆　';
    if (!calm) setInterval(function () {
        title = title.slice(1) + title.charAt(0);
        document.title = title;
    }, 250);

    if (calm) return;
    var stars = ['☆', '★', '✦', '✧', '＊'];
    var colours = ['#ff3399', '#ffcc00', '#33ccff', '#66ff66', '#cc66ff', '#ff6633'];
    var last = 0;
    document.addEventListener('mousemove', function (e) {
        var now = Date.now();
        if (now - last < 40) return;
        last = now;
        var s = document.createElement('span');
        s.className = 'kira';
        s.textContent = stars[Math.random() * stars.length | 0];
        s.style.color = colours[Math.random() * colours.length | 0];
        s.style.left = (e.pageX + 6) + 'px';
        s.style.top = (e.pageY + 6) + 'px';
        s.style.fontSize = (10 + Math.random() * 8 | 0) + 'px';
        document.body.appendChild(s);
        setTimeout(function () { s.remove(); }, 900);
    });
})();
