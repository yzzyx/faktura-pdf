'use strict';

function debounce(fn, delay) {
    let timeoutID;
    return function() {
        var context = this, args = arguments;
        if(timeoutID)
            clearTimeout(timeoutID);
        timeoutID = setTimeout(() => {
            fn.apply(context, args)
        }, delay);
    }
}

// Replace the current customer with an existing or new customer
$("#customer-change").on('click', function () {
    let c = $(this).closest(".card");
    $(".card-display", c).hide();
    $("#customer-controls").show();
});

// Create a new customer entry
$("#customer-new").on('click', function () {
    $("#customer-controls").hide();
    $("#customer-edit").show();
    $("#customer-edit .new-value").removeAttr('disabled');
    $("input[name='customer.id']").val(0); // Clear existing customer connection
    $("#save-btn").show();
});

// Search for existing customers
$("#customer-search").autocomplete({
    source: function(request, response) {
        $.ajax({
            url: "/customer",
            dataType: "json",
            data: {
                search: request.term
            },
            success: function(data) {
                response(data);
            }
        });
    },
    minLength: 2,
    select: function(event, ui) {
        let ctrls = $("#customer-controls");

        ctrls.hide();
        let c = ctrls.closest(".card");
        $(".card-display", c).show();
        $(".card-edit", c).hide();
        $(".card-edit .new-value", c).attr('disabled', true);


        // Populate fields
        for (let n of Object.keys(ui.item)) {
            $(".customer-"+n).text(ui.item[n]);
            $("input[name='customer."+n+"']").val(ui.item[n]);
        }
        $("#save-btn").show();
    },
    focus: function( event, ui ) {
        $("#customer-search").val(ui.item.name);
        return false;
    },
}).autocomplete("instance")._renderItem = function(ul, item) {
    let e = newEl("li", {className: "person",
        children: [
            newEl("span", {className: "name", innerText: item.name}),
            newEl("span", {className: "address", innerText: item.address1 + " " + item.postcode + " " + item.city}),
        ]});
    return $(e).appendTo(ul);
};