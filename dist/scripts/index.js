(function () {
    'use strict';

    function removeClick(el) {
        el.outerHTML = `<div id="${el.id}" class="hidden transition ease-out"></div>`;
    }
    window.removeClick = removeClick;

})();
