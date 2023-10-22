function removeClick(el: any) {
    el.outerHTML = `<div id="${el.id}" class="hidden transition ease-out"></div>`;
}
window.removeClick = removeClick;


