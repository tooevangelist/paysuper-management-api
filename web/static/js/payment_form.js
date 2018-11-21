const PM_ACTIVE_CLASS = 'active';

function getAmount($element) {
    return $element.data('amount') + " " + $element.data('currency')
}

$(function() {
    let $pmItem = $('div.payment-methods div.item');
    let $amountContainer = $('div.details div.amount');
    let $pmFirst = $pmItem.first();

    $('#' + $pmFirst.data('form')).show();
    $amountContainer.html(getAmount($pmFirst));

    $pmItem.on('click', function () {
        $pmItem.removeClass(PM_ACTIVE_CLASS);
        $(this).addClass(PM_ACTIVE_CLASS);

        $('div.payment-method-requisites').find('div.form').hide();
        $('#' + $(this).data('form')).show();
        $amountContainer.html(getAmount($(this)));
    });

    $('input.number').on('keyup', function () {
        $(this).val($(this).val().replace(/\D/g,''));
    });

    $('form#order-form').on('submit', function (e) {
        e.preventDefault();

        let data = new FormData($(this)[0]);
        let object = {};

        data.forEach(function(value, key){
            object[key] = value;
        });

        $.ajax({
            url: '/api/v1/payment',
            type: 'post',
            data: JSON.stringify(object),
            success: function(data) {
                console.log(data);
            },
            error: function(xhr) {
                let error = JSON.parse(xhr['responseText']);
                console.log(error);
            },
            cache: false,
            contentType: 'application/json; charset=utf-8',
            processData: false,
            dataType: 'json'
        });
    });
});