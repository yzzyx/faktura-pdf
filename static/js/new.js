'use strict';

// newEl is a helper function to simplify creation of elements
//  expects 'args' to be an object, that is used to populate the attributes of the created element
function newEl(type, args) {
    let el = document.createElement(type);
    if (args !== undefined) {
        let children = args["children"];
        let attributes = args["attributes"];
        delete(args["children"]);
        delete(args["attributes"]);
        Object.assign(el, args);
        if (children !== undefined) {
            el.append(...children);
        }
        if (attributes !== undefined) {
            let attrNames = Object.keys(attributes);
            for (let i = 0; i < attrNames.length; i++) {
                el.setAttribute(attrNames[i], attributes[attrNames[i]]);
            }
        }
    }
    return el;
}
