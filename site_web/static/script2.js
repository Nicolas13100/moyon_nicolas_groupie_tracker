document.addEventListener('DOMContentLoaded', function() {
    // Get the image in game 3
    var game3Img = document.querySelector('.game3 img');

    // Function to set max height after image loads
    function setMaxHeight() {
        // Get the height of the image in game 3
        var img3Height = game3Img.clientHeight;

        // Calculate half of the height of the image in game 3
        var maxImgHeight = Math.floor((img3Height / 2)-7);

        // Set the max-height of all game images to be half the height of the image in game 3
        var gameImages = document.querySelectorAll('.game1 img, .game2 img, .game4 img, .game5 img, .game6 img');
        gameImages.forEach(function(img) {
            img.style.maxHeight = maxImgHeight + 'px';
        });
    }

    // Check if the image is available
    if (game3Img) {
        // If the image is already loaded, set the max height immediately
        if (game3Img.complete) {
            setMaxHeight();
        } else {
            // Otherwise, wait for the image to load
            game3Img.onload = setMaxHeight;
        }
    } else {
        console.error("Image in game 3 not found.");
    }
});