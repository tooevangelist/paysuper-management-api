const PM_ACTIVE_CLASS = 'active';

function getAmount($element) {
    return $element.data('amount') + " " + $element.data('currency')
}

$(function() {
    let $pmItem = $('div.payment-methods div.item');
    let $amountContainer = $('div.details div.amount');
    let $selectedPmInput = $('form#order-form input[name=payment_method_id]');

    $pmItem.each(function (i, el) {
        if (i === 0) {
            $('#' + $(el).data('form')).show();
        } else {
            $('#' + $(el).data('form')).find('input').attr({disabled: 'disabled'});
        }
    });

    if ($.trim($amountContainer.html()).length <= 0) {
        $amountContainer.html(getAmount($pmItem.first()));
    }

    if ($selectedPmInput.val().length <= 0) {
        $selectedPmInput.val($pmItem.first().data('identifier'));
    }

    $pmItem.on('click', function () {
        $pmItem.removeClass(PM_ACTIVE_CLASS);
        $(this).addClass(PM_ACTIVE_CLASS);

        let $hideEl = $('div.payment-method-requisites').find('div.form');

        $hideEl.hide();
        $hideEl.find('input').attr({disabled: 'disabled'});

        let $showEl = $('#' + $(this).data('form'));

        $showEl.show();
        $showEl.find('input').removeAttr('disabled');

        $amountContainer.html(getAmount($(this)));
        $selectedPmInput.val($(this).data('identifier'));
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