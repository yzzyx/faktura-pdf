'use strict';
$(function () {
    $(".edit").on('click', function (ev) {
        let c = $(this).closest(".card");
        $(".card-display", c).hide();
        $(".card-edit", c).show();
        $(".card-edit .new-value", c).removeAttr('disabled');
        $("#save-btn").show();
    });
});
