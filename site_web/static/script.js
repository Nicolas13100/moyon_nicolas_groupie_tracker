document.addEventListener('DOMContentLoaded', function () {
    var isMouseDown = false;
    var startX, scrollLeft;

    var indexTop = document.querySelector('.indexTop');

    indexTop.addEventListener('mousedown', function (e) {
        e.preventDefault(); // Prevent text selection
        isMouseDown = true;
        startX = e.pageX - indexTop.offsetLeft;
        scrollLeft = indexTop.scrollLeft;
    });

    indexTop.addEventListener('mouseup', function () {
        isMouseDown = false;
    });

    indexTop.addEventListener('mouseleave', function () {
        isMouseDown = false;
    });

    indexTop.addEventListener('mousemove', function (e) {
        if (!isMouseDown) return;
        var x = e.pageX - indexTop.offsetLeft;
        var walk = (x - startX) * 1; // Adjust the multiplier for smoother scrolling

        var maxScrollLeft = indexTop.scrollWidth - indexTop.clientWidth;

        indexTop.scrollLeft = Math.max(0, Math.min(scrollLeft - walk, maxScrollLeft));
    });
});
document.addEventListener('DOMContentLoaded', function () {
    // Get a reference to the login button
    var loginButton = document.getElementById('loginButton');

    // Add a click event listener to the login button
    loginButton.addEventListener('click', function () {
        // Redirect to the login page when the button is clicked
        window.location.href = "/login";
    });
});