'use strict';
$(function () {
    let newRowID = -1;

    $(".datepicker").datepicker({
        dateFormat: "yy-mm-dd",
        firstDay: 1,
        dayNames: [ "Söndag", "Måndag", "Tisdag", "Onsdag", "Torsdag", "Fredag", "Lördag" ],
        dayNamesMin: [ "Sö", "Må", "Ti", "On", "To", "Fr", "Lö" ],
        dayNamesShort: [ "Sön", "Mån", "Tis", "Ons", "Tor", "Fre", "Lör" ],
        monthNames: [ "Januari", "Februari", "Mars", "April", "Maj", "Juni", "Juli", "Augusti", "September", "Oktober", "November", "December" ],
        monthNamesShort: [ "Jan", "Feb", "Mar", "Apr", "Maj", "Jun", "Jul", "Aug", "Sep", "Okt", "Nov", "Dec" ]
    });

    let rows = $("#invoice_rows tbody");
    if (!rows.data("disabled")) {
        $("#invoice_rows tbody").sortable({
            "axis": "y",
            "cancel": ".edit",
            "stop": function () {
                let x = document.querySelectorAll("#invoice_rows tbody tr");
                let roworder = [];
                for (let idx = 0; idx < x.length; idx++) {
                    let id = parseInt(x[idx].dataset["row"]);

                    // For rows that we still haven't created, we need to set the row index directly
                    if (isNaN(id)) {
                        let input = $("input[name='row[]']", x[idx]);
                        let row = JSON.parse(input.val());
                        row.row_order = idx;
                        input.val(JSON.stringify(row));
                        id = 0; // placeholder for the new id
                    }
                    roworder.push(id);
                }
                $("input[name=roworder]").val(JSON.stringify(roworder));
                $("#save-btn").show();
            }
        });
    }

    $(".card .edit").on('click', function (ev) {
        let c = $(this).closest(".card");
        $(".card-display", c).hide();
        $(".card-edit", c).show();
        $(".card-edit .new-value", c).removeAttr('disabled');
        $("#save-btn").show();
    });

    $("#invoice-show-add-row").on('click', function (ev) {
        $("#invoice-row-description").val("");
        $("#invoice-row-id").val(0);
        $(".invoice-row-update").hide();

        // Only show ROT/RUT info if the invoice accepts it
        let rutApplicable = $("input[name='rut_applicable']").is(":checked");
        if (rutApplicable) {
            $("#invoice-row-rut-rot").prop("disabled", false);
        } else {
            $("#invoice-row-rut-rot").prop("disabled", true)
                .prop("checked", false);
        }

        $(".invoice-row-add").show();
        $('#invoice-row-modal').modal('show');
    });

    $("#invoice_rows").on('click', "tbody tr .edit", function (ev) {
        let tr = $(this).closest("tr");
        let v = tr.data("json");
        if (!v) {
            v = JSON.parse($("input[name='row[]']", tr).val());
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

        // Only show ROT/RUT info if the invoice accepts it
        let isROTRUT = v.is_rot_rut;
        let rutApplicable = $("input[name='rut_applicable']").is(":checked");
        if (rutApplicable) {
            $("#invoice-row-rut-rot").prop("disabled", false);
        } else {
            $("#invoice-row-rut-rot").prop("disabled", true);
            isROTRUT = false;
        }

        // Make sure we don't change dropdowns
        lastType = $("input[name='invoice-row-rut-rot-type']:checked").val();
        $("#invoice-row-rut-rot").prop("checked", isROTRUT);
        if (isROTRUT) {
            $("#invoice-row-rut-rot-type-rut").prop("checked", v.rot_rut_service_type > 6);
            $("#invoice-row-rut-rot-type-rot").prop("checked", v.rot_rut_service_type < 7);
        }
        update_price();

        $('#invoice-row-modal').modal('show');
    });

    $("#delete-row").on('click', function (ev) {
        let id = parseInt($("#invoice-row-id").val());

        if (id > 0) {
            let el = newEl("input", {name: "delete_row[]", type: "hidden", value: id});
            $("#invoice_rows tbody tr[data-row=" + id +"]").replaceWith(el);
        } else {
            $("#invoice_rows tbody tr[data-row=" + id +"]").remove();
        }
        $('#invoice-row-modal').modal('hide');
        $("#save-btn").show();
    });

    let vatAmounts = {
        0: 0.25,
        1: 0.12,
        2: 0.06,
        3: 0
    }

    function update_totals(ev) {
        let totalIncl = 0;
        let totalCustomer = 0;
        let totalVAT25 = 0;
        let totalVAT12 = 0;
        let totalVAT6 = 0;
        let totalROTRUT = 0;

        let rows = $("#invoice_rows tbody tr");
        for (let idx = 0; idx < rows.length; idx++) {
            let tr = $(rows[idx]);
            let v = tr.data("json");
            if (!v) {
                v = JSON.parse($("input[name='row[]']", tr).val());
            }

            let rowTotal = v.cost * v.count;
            let priceInclRUT = v.cost;
            let isROTRUT = v.is_rot_rut;

            let rutApplicable = $("input[name='rut_applicable']").is(":checked");
            if (!rutApplicable) {
                isROTRUT = false;
            }

            if (isROTRUT && v.rot_rut_service_type < 7) { // ROT
                priceInclRUT = v.cost * 0.7;
                totalROTRUT = totalROTRUT + (v.cost * 0.3 * v.count)
            } else if (isROTRUT && v.rot_rut_service_type > 6) { // RUT
                priceInclRUT = v.cost * 0.5;
                totalROTRUT = totalROTRUT + (v.cost * 0.5 * v.count)
            }

            totalIncl = totalIncl + rowTotal;
            totalCustomer = totalCustomer + priceInclRUT * v.count;

            let vatAmount = rowTotal - rowTotal/(1+vatAmounts[v.vat]);
            switch (v.vat) {
                case 0:
                    totalVAT25 = totalVAT25 + vatAmount
                    break;
                case 1:
                    totalVAT12 = totalVAT12 + vatAmount
                    break;
                case 2:
                    totalVAT6 = totalVAT6 + vatAmount
                    break;
            }
        }

        $("#total-incl .sum").text(totalIncl.toFixed(2)).parent().toggle(totalIncl>0);
        $("#total-vat-25 .sum").text(totalVAT25.toFixed(2)).parent().toggle(totalVAT25>0);
        $("#total-vat-12 .sum").text(totalVAT12.toFixed(2)).parent().toggle(totalVAT12>0);
        $("#total-vat-6 .sum").text(totalVAT6.toFixed(2)).parent().toggle(totalVAT6>0);
        $("#total-rot-rut .sum").text(totalROTRUT.toFixed(2)).parent().toggle(totalROTRUT>0);
        $("#total-customer .sum").text(totalCustomer.toFixed(2)).parent().toggle(totalCustomer>0);
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
    $("input[name='rut_applicable']").on({change: update_totals});

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
            row_order: document.querySelectorAll("#invoice_rows tbody tr").length,
        };

        let id = $("#invoice-row-id").val()
        let isNewEntry = false;
        entry.ID = parseInt(id);
        if (entry.ID === 0) {
            entry.ID = newRowID--;
            isNewEntry = true;
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
                newEl("td", {classList: "text-center", children: [ newEl("i", {classList: "fa fa-chevron-right edit"}), ]})
            ]
        });

        el.dataset["row"] = entry.ID;
        if (isNewEntry) {
            $("#invoice_rows tbody").append(el);
        } else {
            $("#invoice_rows tbody tr[data-row=" + entry.ID +"]").replaceWith(el);
        }
        $("#invoice-row-modal").modal('hide');
        $("#save-btn").show();
        update_totals();
        return false;
    });

    $("input[name='customer.pnr']").keyup(function () {
        if (!/\d{12}/.test($(this).val())) {
            $("#customer-pnr-error").show();
        } else {
            $("#customer-pnr-error").hide();
        }
    });
});
