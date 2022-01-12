
// from https://stackoverflow.com/questions/3710204
/**
 * If you don't care about primitives and only objects then this function
 * is for you, otherwise look elsewhere.
 * This function will return `false` for any valid json primitive.
 * EG, 'true' -> false
 *     '123' -> false
 *     'null' -> false
 *     '"I'm a string"' -> false
 */
function tryParseJSONObject (jsonString){
    try {
        var o = JSON.parse(jsonString);

        // Handle non-exception-throwing cases:
        // Neither JSON.parse(false) or JSON.parse(1234) throw errors, hence the type-checking,
        // but... JSON.parse(null) returns null, and typeof null === "object",
        // so we must check for that, too. Thankfully, null is falsey, so this suffices:
        if (o && typeof o === "object") {
            return o;
        }
    } catch (e) { 
        return false; 
    }

    return false;
};

function verifyJSObjectStructure(obj, validObj) {

    if (!Array.isArray(validObj)) {
        return false;
    }

    // get from https://stackoverflow.com/questions/48156589
    let validate = (obj,props) => props.every(prop => obj.hasOwnProperty(prop));
    if (validate(obj, validObj)){
        return true;
    }

    return false;
}

function getJSObjFromJsonString (jsonString, validObj) {

    obj = tryParseJSONObject (jsonString);
    if (obj === false) {
        return false;
    }

    if (!verifyJSObjectStructure(obj, validObj)) {
        return false;
    }

    return obj;
}

