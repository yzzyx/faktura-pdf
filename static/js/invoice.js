'use strict';
$(function () {
    $(".datepicker").datepicker({
        dateFormat: "yy-mm-dd",
        firstDay: 1,
        dayNames: [ "Söndag", "Måndag", "Tisdag", "Onsdag", "Torsdag", "Fredag", "Lördag" ],
        dayNamesMin: [ "Sö", "Må", "Ti", "On", "To", "Fr", "Lö" ],
        dayNamesShort: [ "Sön", "Mån", "Tis", "Ons", "Tor", "Fre", "Lör" ],
        monthNames: [ "Januari", "Februari", "Mars", "April", "Maj", "Juni", "Juli", "Augusti", "September", "Oktober", "November", "December" ],
        monthNamesShort: [ "Jan", "Feb", "Mar", "Apr", "Maj", "Jun", "Jul", "Aug", "Sep", "Okt", "Nov", "Dec" ]
    });

    $(".edit").on('click', function (ev) {
        $(this).prevAll(".current-value").hide();
        $(this).prevAll(".new-value").removeAttr('disabled').show();
        $(this).hide();
    });

    $("#invoice-show-add-row").on('click', function (ev) {
        $("#invoice-row-description").val("");
        $("#invoice-row-id").val(0);
        $(".invoice-row-update").hide();
        $(".invoice-row-add").show();
        $('#invoice-row-modal').modal('show');
    });

    $("#invoice_rows").on('click', "tbody tr", function (ev) {
        let tr = $(this);
        let v = tr.data("json");
        if (!v) {
            v = JSON.parse($("input[name='row[]']").val());
        }

        $(".invoice-row-update").show();
        $(".invoice-row-add").hide();

        $("#invoice-row-id").val(v.ID);
        $("#invoice-row-description").val(v.description);
        $("#invoice-row-price-incl").val(v.cost);
        $("#invoice-row-count").val(v.count);
        $("#invoice-row-unit").val(v.unit);
        $("#invoice-row-vat").val(v.vat);

        $("#invoice-row-rut-rot-service-type").val(v.rot_rut_service_type);
        // Make sure we don't change dropdowns
        lastType = $("input[name='invoice-row-rut-rot-type']:checked").val();
        $("#invoice-row-rut-rot").prop("checked", v.is_rot_rut);
        if (v.is_rot_rut) {
            $("#invoice-row-rut-rot-type-rut").prop("checked", v.rot_rut_service_type > 6);
            $("#invoice-row-rut-rot-type-rot").prop("checked", v.rot_rut_service_type < 7);
        }
        update_price();

        $('#invoice-row-modal').modal('show');
    });

    $("#delete-row").on('click', function (ev) {
        let id = parseInt($("#invoice-row-id").val());
        let el = newEl("input", {name: "delete_row[]", type: "hidden", value: id});
        $("#invoice_rows tbody tr[data-row=" + id +"]").replaceWith(el);
        $('#invoice-row-modal').modal('hide');
    });

    let vatAmounts = {
        0: 0.25,
        1: 0.12,
        2: 0.06,
        3: 0
    }

    let update_price = function (ev) {
        let $t = $(this);
        if (!ev || this.id === "invoice-row-vat") {
            $t = $("#invoice-row-price-incl");
        }

        let price = parseFloat($t.val());
        let vat = vatAmounts[parseInt($("#invoice-row-vat").val())];

        if (price === undefined) return;

        if (!$t.data("vat-incl")) {
            price = price*(1+vat);
        } else {
            price = price/(1+vat);
        }

        if (parseInt(price) !== price) price = price.toFixed(2);
        $($t.data("price-target")).val(price);
    };

    $("#invoice-row-price-incl").on({keyup: update_price, change: update_price});
    $("#invoice-row-price-exkl").on({keyup: update_price, change: update_price});
    $("#invoice-row-vat").on({change: update_price});

    let lastType = "";
    let showRutRotService = function () {
        let el = $(".show-rut-rot");
        if ($("#invoice-row-rut-rot").is(":checked")) {
            el.show();
        } else {
            el.hide();
        }

        let type = $("input[name='invoice-row-rut-rot-type']:checked").val();
        let optRUT = $("#invoice-row-rut-rot-service-type option.rut");
        let optROT = $("#invoice-row-rut-rot-service-type option.rot");
        if (type === "rut") {
            optRUT.show().addClass("active");
            optROT.hide().removeClass("active");
        } else {
            optRUT.hide().removeClass("active");
            optROT.show().addClass("active");
        }

        if (type !== lastType) {
            lastType = type;
            $("#invoice-row-rut-rot-service-type option.active.default").prop("selected","selected");
        }
    };

    $("#invoice-row-rut-rot").on({change: showRutRotService});
    $("input[name='invoice-row-rut-rot-type']").on({change: showRutRotService});
    $('#invoice-row-modal').on('show.bs.modal', showRutRotService);

    $("#add-row").on('click', function (ev) {
        let form = $("#invoice-row-form")[0];
        if (form.checkValidity() === false) {
            ev.preventDefault();
            ev.stopPropagation();
            form.classList.add("was-validated");
            return false;
        }

        let entry = {
            description: $("#invoice-row-description").val(),
            cost: parseFloat($("#invoice-row-price-incl").val()),
            count: parseFloat($("#invoice-row-count").val()),
            unit: parseInt($("#invoice-row-unit").val()),
            vat: parseInt($("#invoice-row-vat").val()),
            is_rot_rut: $("#invoice-row-rut-rot").is(":checked"),
        };

        let id = $("#invoice-row-id").val()
        if (id) {
            entry.ID = parseInt(id);
        }

        if (entry.is_rot_rut) {
            entry.rot_rut_service_type = parseInt($("#invoice-row-rut-rot-service-type").val());
        }
        entry.total = entry.cost * entry.count;


        let units = ["-", "st", "timmar", "dagar"];
        let vat = ["25 %", "12 %", "6 %", "0 %"];
        let countText =  entry.count;
        if (entry.unit > 0) {
            countText += " " + units[entry.unit];
        }

        let el = newEl("tr", { classList: "new",
            children: [
                newEl("td", {
                    children: [
                        newEl("input", {name: "row[]", type: "hidden", value: JSON.stringify(entry)}),
                        newEl("span", {textContent: entry.description}),
                    ]
                }),
                newEl("td", {classList: "text-right", children: [ newEl("span", {textContent: entry.cost.toFixed(2)}), ]}),
                newEl("td", {classList: "text-right", children: [ newEl("span", {textContent: countText}), ]}),
                newEl("td", {classList: "text-right", children: [ newEl("span", {textContent: entry.total.toFixed(2)}), ]}),
                newEl("td", {classList: "text-right", children: [ newEl("span", {textContent: vat[entry.vat]}), ]}),
                newEl("td", {classList: "text-center", children: [ entry.is_rot_rut ? newEl("i", {classList: "fa fa-check"}) : "", ]}),
                newEl("td", {classList: "text-center", children: [ newEl("i", {classList: "fa fa-chevron-right"}), ]})
            ]
        });

        if (entry.ID) {
            el.dataset["row"] = entry.ID;
            $("#invoice_rows tbody tr[data-row=" + entry.ID +"]").replaceWith(el);
        } else {
            $("#invoice_rows tbody").append(el);
        }
        $("#invoice-row-modal").modal('hide');
        return false;
    });
});
