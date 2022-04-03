/* Make tables sortable */
$("table.sortable").each(function () {
    let $table = $(this);

    $("th", $table).on("click", function (ev) {
        let target = $(ev.target).closest("th")
        let column = $(target).data("column");

        // Toggle direction
        let dir = "desc";
        if ($(target).hasClass("sort-desc")) {
            dir = "asc";
        }

        if ($table.hasClass("multisort")) {
            if (!$(target).is('[class^="sort-"]')) {
                $(target).addClass("sort-" + dir);
            } else {
                if ($(target.hasClass("sort-desc"))) {
                    $(target).toggleClass("sort-asc sort-desc")
                } else if ($(target.hasClass("sort-asc"))) {
                    $(target).toggleClass("sort-desc sort-asc")
                }
            }

            let orderList = $table.data("orderby");
            if (orderList) {
                $table.data("orderby", addOrdering(orderList, column, dir, $("th", $table).length));
            } else {
                $table.data("orderby", [column + "_" + dir]);
            }
        } else {
            $table.data("orderby", column)
            $("th", $table).removeClass("sort-asc sort-desc fa-sort-amount-down fa-sort-amount-up");
            $(target).addClass("sort-" + dir);
        }

        $table.data("dir", dir);
        loadTableContents($table);
    });
});

/* Add sorting icons */
$("table.sortable").click(function(e) {
    let target = $(e.target).closest("th")
    $(this).find("th").each(function () {
        if (!$(this).hasClass("sort-desc") && !$(this).hasClass("sort-asc")) {
            $(this).find("span").removeClass("fa-sort-amount-up").addClass("fa-sort-amount-down").css("color", "lightgrey");
        }
    });
    let $span = $(target).find("span");
    if ($span.css("color") !== "#e8e8e8") {
        $span.css("color", "black")
    }
    $span.toggleClass("fa-sort-amount-up fa-sort-amount-down");
});

/* Maintain a list of names with trailing direction, leftmost is last clicked */
function addOrdering(list, name, dir, maxlength) {
    let delim = "_"
    let cleanList = list.filter((elem) => !elem.startsWith(name + delim))
    name += delim + dir
    let newList = [name, ...cleanList]
    return newList.slice(0,maxlength)
}
