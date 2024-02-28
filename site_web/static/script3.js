function fetchData(page) {
    // Make AJAX request to fetch data for the specified page.
    // Example using Fetch API:
    fetch('/pagination?page=' + page)
        .then(response => response.json())
        .then(data => {
            // Update UI with fetched data.
        })
        .catch(error => {
            console.error('Error fetching data:', error);
        });
}

function renderPagination(totalPages) {
    // Generate pagination controls based on the total number of pages.
    // Example: display page numbers and attach click event handlers.
}

// Initial page load
fetchData(1);

// Example: Render pagination controls when the page is loaded.
renderPagination(totalPages);

// Example: Event listener for clicking on a page number.
document.getElementById('pagination').addEventListener('click', function (event) {
    if (event.target.tagName === 'A') {
        const page = parseInt(event.target.dataset.page);
        fetchData(page);
    }
});