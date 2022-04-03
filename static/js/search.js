/* Add search-functionality to tables etc. */
$(".search").each(function() {
    let $input = $(this);
    let $table = $($input.data("table"));

    $(this).on("input", debounce(function (ev) {
        // If target is set to a table, we should only update part of that table (e.g. tbody),
        // which is specified in the 'data-target' property
        $table.data("search", $input.val());
        loadTableContents($table);
    }, 250));
});

let currentRequest;
// loadTableContents fetches the contents of a specific table based on the data-parameters of that table:
//   data-url: which URL to fetch from
//   data-target: where to place the result of the request
//   data-*: all other data parameters will be automatically added to the request as parameters
function loadTableContents($table) {

    let $target = $($table.data("target"));
    let parameters = {};
    let data = $table.data();
    for (let key in data) {
        if (key !== "url" && key !== "target" && key !== "page" && data[key] !== "") {
            parameters[key] = data[key];
        }
    }

    // If we already have a request, cancel it
    if (currentRequest && currentRequest.readyState !== 4) {
        currentRequest.abort();
    }

    let url = $table.data("url");
    if (url === undefined) {
        url = "";
    }

    // Update URL so that we can reload page and still show the same result
    let u = location.protocol + "//" + location.host + location.pathname;
    if (Object.keys(parameters).length > 0) {
        u += "?" + $.param(parameters);
    }

    history.pushState(null,"", u);

    parameters["content"] = 1;
    currentRequest = $.get(url + "?" + $.param(parameters),
        function (data, textstatus, xhr) {
            $target.html(data);
        });
}
