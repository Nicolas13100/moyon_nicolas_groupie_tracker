function submitForm() {
    var form = document.getElementById("tagForm");
    form.action = window.location.href;
    form.submit();
}

document.addEventListener("DOMContentLoaded", function () { 
    // Function to get the value of a URL parameter by name
    function getParameterByName(name, url) {
        if (!url) url = window.location.href;
        name = name.replace(/[\[\]]/g, "\\$&");
        var regex = new RegExp("[?&]" + name + "(=([^&#]*)|&|#|$)"),
            results = regex.exec(url);
        if (!results) return null;
        if (!results[2]) return '';
        return decodeURIComponent(results[2].replace(/\+/g, " "));
    }

    // Get the query parameter from the current URL
    var query = getParameterByName('query');

    function generatePagination(totalPages, currentPage) {
        var paginationContainer = document.getElementById("pagination");
        paginationContainer.innerHTML = ""; // Clear previous pagination links

        for (var i = 1; i <= totalPages; i++) {
            var li = document.createElement("li");
            li.classList.add("page-item");
            if (i === currentPage) {
                li.classList.add("active");
            }
            var a = document.createElement("a");
            a.classList.add("page-link");
            a.href = "?query=" + query + "&page=" + i;
            a.textContent = i;
            li.appendChild(a);
            paginationContainer.appendChild(li);
        }
    }
    // Call the function with total pages and current page number
    generatePagination(totalPages, currentPage);
});