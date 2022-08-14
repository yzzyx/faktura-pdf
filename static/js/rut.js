'use strict';
$(function () {
    $("input").change(function (){
        $("#save-btn").show();
    }).keyup(function (){
        $("#save-btn").show();
    });
});
