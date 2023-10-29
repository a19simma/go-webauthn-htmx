(function () {
    'use strict';

    /******************************************************************************
    Copyright (c) Microsoft Corporation.

    Permission to use, copy, modify, and/or distribute this software for any
    purpose with or without fee is hereby granted.

    THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
    REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
    AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
    INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
    LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
    OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
    PERFORMANCE OF THIS SOFTWARE.
    ***************************************************************************** */
    /* global Reflect, Promise, SuppressedError, Symbol */


    function __awaiter(thisArg, _arguments, P, generator) {
        function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
        return new (P || (P = Promise))(function (resolve, reject) {
            function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
            function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
            function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
            step((generator = generator.apply(thisArg, _arguments || [])).next());
        });
    }

    typeof SuppressedError === "function" ? SuppressedError : function (error, suppressed, message) {
        var e = new Error(message);
        return e.name = "SuppressedError", e.error = error, e.suppressed = suppressed, e;
    };

    /* [@simplewebauthn/browser@8.2.1] */
    function utf8StringToBuffer(value) {
        return new TextEncoder().encode(value);
    }

    function bufferToBase64URLString(buffer) {
        const bytes = new Uint8Array(buffer);
        let str = '';
        for (const charCode of bytes) {
            str += String.fromCharCode(charCode);
        }
        const base64String = btoa(str);
        return base64String.replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
    }

    function base64URLStringToBuffer(base64URLString) {
        const base64 = base64URLString.replace(/-/g, '+').replace(/_/g, '/');
        const padLength = (4 - (base64.length % 4)) % 4;
        const padded = base64.padEnd(base64.length + padLength, '=');
        const binary = atob(padded);
        const buffer = new ArrayBuffer(binary.length);
        const bytes = new Uint8Array(buffer);
        for (let i = 0; i < binary.length; i++) {
            bytes[i] = binary.charCodeAt(i);
        }
        return buffer;
    }

    function browserSupportsWebAuthn() {
        return (window?.PublicKeyCredential !== undefined &&
            typeof window.PublicKeyCredential === 'function');
    }

    function toPublicKeyCredentialDescriptor(descriptor) {
        const { id } = descriptor;
        return {
            ...descriptor,
            id: base64URLStringToBuffer(id),
            transports: descriptor.transports,
        };
    }

    function isValidDomain(hostname) {
        return (hostname === 'localhost' ||
            /^([a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,}$/i.test(hostname));
    }

    class WebAuthnError extends Error {
        constructor({ message, code, cause, name, }) {
            super(message, { cause });
            this.name = name ?? cause.name;
            this.code = code;
        }
    }

    function identifyRegistrationError({ error, options, }) {
        const { publicKey } = options;
        if (!publicKey) {
            throw Error('options was missing required publicKey property');
        }
        if (error.name === 'AbortError') {
            if (options.signal instanceof AbortSignal) {
                return new WebAuthnError({
                    message: 'Registration ceremony was sent an abort signal',
                    code: 'ERROR_CEREMONY_ABORTED',
                    cause: error,
                });
            }
        }
        else if (error.name === 'ConstraintError') {
            if (publicKey.authenticatorSelection?.requireResidentKey === true) {
                return new WebAuthnError({
                    message: 'Discoverable credentials were required but no available authenticator supported it',
                    code: 'ERROR_AUTHENTICATOR_MISSING_DISCOVERABLE_CREDENTIAL_SUPPORT',
                    cause: error,
                });
            }
            else if (publicKey.authenticatorSelection?.userVerification === 'required') {
                return new WebAuthnError({
                    message: 'User verification was required but no available authenticator supported it',
                    code: 'ERROR_AUTHENTICATOR_MISSING_USER_VERIFICATION_SUPPORT',
                    cause: error,
                });
            }
        }
        else if (error.name === 'InvalidStateError') {
            return new WebAuthnError({
                message: 'The authenticator was previously registered',
                code: 'ERROR_AUTHENTICATOR_PREVIOUSLY_REGISTERED',
                cause: error,
            });
        }
        else if (error.name === 'NotAllowedError') {
            return new WebAuthnError({
                message: error.message,
                code: 'ERROR_PASSTHROUGH_SEE_CAUSE_PROPERTY',
                cause: error,
            });
        }
        else if (error.name === 'NotSupportedError') {
            const validPubKeyCredParams = publicKey.pubKeyCredParams.filter((param) => param.type === 'public-key');
            if (validPubKeyCredParams.length === 0) {
                return new WebAuthnError({
                    message: 'No entry in pubKeyCredParams was of type "public-key"',
                    code: 'ERROR_MALFORMED_PUBKEYCREDPARAMS',
                    cause: error,
                });
            }
            return new WebAuthnError({
                message: 'No available authenticator supported any of the specified pubKeyCredParams algorithms',
                code: 'ERROR_AUTHENTICATOR_NO_SUPPORTED_PUBKEYCREDPARAMS_ALG',
                cause: error,
            });
        }
        else if (error.name === 'SecurityError') {
            const effectiveDomain = window.location.hostname;
            if (!isValidDomain(effectiveDomain)) {
                return new WebAuthnError({
                    message: `${window.location.hostname} is an invalid domain`,
                    code: 'ERROR_INVALID_DOMAIN',
                    cause: error,
                });
            }
            else if (publicKey.rp.id !== effectiveDomain) {
                return new WebAuthnError({
                    message: `The RP ID "${publicKey.rp.id}" is invalid for this domain`,
                    code: 'ERROR_INVALID_RP_ID',
                    cause: error,
                });
            }
        }
        else if (error.name === 'TypeError') {
            if (publicKey.user.id.byteLength < 1 || publicKey.user.id.byteLength > 64) {
                return new WebAuthnError({
                    message: 'User ID was not between 1 and 64 characters',
                    code: 'ERROR_INVALID_USER_ID_LENGTH',
                    cause: error,
                });
            }
        }
        else if (error.name === 'UnknownError') {
            return new WebAuthnError({
                message: 'The authenticator was unable to process the specified options, or could not create a new credential',
                code: 'ERROR_AUTHENTICATOR_GENERAL_ERROR',
                cause: error,
            });
        }
        return error;
    }

    class WebAuthnAbortService {
        createNewAbortSignal() {
            if (this.controller) {
                const abortError = new Error('Cancelling existing WebAuthn API call for new one');
                abortError.name = 'AbortError';
                this.controller.abort(abortError);
            }
            const newController = new AbortController();
            this.controller = newController;
            return newController.signal;
        }
    }
    const webauthnAbortService = new WebAuthnAbortService();

    const attachments = ['cross-platform', 'platform'];
    function toAuthenticatorAttachment(attachment) {
        if (!attachment) {
            return;
        }
        if (attachments.indexOf(attachment) < 0) {
            return;
        }
        return attachment;
    }

    async function startRegistration(creationOptionsJSON) {
        if (!browserSupportsWebAuthn()) {
            throw new Error('WebAuthn is not supported in this browser');
        }
        const publicKey = {
            ...creationOptionsJSON,
            challenge: base64URLStringToBuffer(creationOptionsJSON.challenge),
            user: {
                ...creationOptionsJSON.user,
                id: utf8StringToBuffer(creationOptionsJSON.user.id),
            },
            excludeCredentials: creationOptionsJSON.excludeCredentials?.map(toPublicKeyCredentialDescriptor),
        };
        const options = { publicKey };
        options.signal = webauthnAbortService.createNewAbortSignal();
        let credential;
        try {
            credential = (await navigator.credentials.create(options));
        }
        catch (err) {
            throw identifyRegistrationError({ error: err, options });
        }
        if (!credential) {
            throw new Error('Registration was not completed');
        }
        const { id, rawId, response, type } = credential;
        let transports = undefined;
        if (typeof response.getTransports === 'function') {
            transports = response.getTransports();
        }
        let responsePublicKeyAlgorithm = undefined;
        if (typeof response.getPublicKeyAlgorithm === 'function') {
            try {
                responsePublicKeyAlgorithm = response.getPublicKeyAlgorithm();
            }
            catch (error) {
                warnOnBrokenImplementation('getPublicKeyAlgorithm()', error);
            }
        }
        let responsePublicKey = undefined;
        if (typeof response.getPublicKey === 'function') {
            try {
                const _publicKey = response.getPublicKey();
                if (_publicKey !== null) {
                    responsePublicKey = bufferToBase64URLString(_publicKey);
                }
            }
            catch (error) {
                warnOnBrokenImplementation('getPublicKey()', error);
            }
        }
        let responseAuthenticatorData;
        if (typeof response.getAuthenticatorData === 'function') {
            try {
                responseAuthenticatorData = bufferToBase64URLString(response.getAuthenticatorData());
            }
            catch (error) {
                warnOnBrokenImplementation('getAuthenticatorData()', error);
            }
        }
        return {
            id,
            rawId: bufferToBase64URLString(rawId),
            response: {
                attestationObject: bufferToBase64URLString(response.attestationObject),
                clientDataJSON: bufferToBase64URLString(response.clientDataJSON),
                transports,
                publicKeyAlgorithm: responsePublicKeyAlgorithm,
                publicKey: responsePublicKey,
                authenticatorData: responseAuthenticatorData,
            },
            type,
            clientExtensionResults: credential.getClientExtensionResults(),
            authenticatorAttachment: toAuthenticatorAttachment(credential.authenticatorAttachment),
        };
    }
    function warnOnBrokenImplementation(methodName, cause) {
        console.warn(`The browser extension that intercepted this WebAuthn API call incorrectly implemented ${methodName}. You should report this error to them.\n`, cause);
    }

    function bufferToUTF8String(value) {
        return new TextDecoder('utf-8').decode(value);
    }

    function browserSupportsWebAuthnAutofill() {
        const globalPublicKeyCredential = window
            .PublicKeyCredential;
        if (globalPublicKeyCredential.isConditionalMediationAvailable === undefined) {
            return new Promise((resolve) => resolve(false));
        }
        return globalPublicKeyCredential.isConditionalMediationAvailable();
    }

    function identifyAuthenticationError({ error, options, }) {
        const { publicKey } = options;
        if (!publicKey) {
            throw Error('options was missing required publicKey property');
        }
        if (error.name === 'AbortError') {
            if (options.signal instanceof AbortSignal) {
                return new WebAuthnError({
                    message: 'Authentication ceremony was sent an abort signal',
                    code: 'ERROR_CEREMONY_ABORTED',
                    cause: error,
                });
            }
        }
        else if (error.name === 'NotAllowedError') {
            return new WebAuthnError({
                message: error.message,
                code: 'ERROR_PASSTHROUGH_SEE_CAUSE_PROPERTY',
                cause: error,
            });
        }
        else if (error.name === 'SecurityError') {
            const effectiveDomain = window.location.hostname;
            if (!isValidDomain(effectiveDomain)) {
                return new WebAuthnError({
                    message: `${window.location.hostname} is an invalid domain`,
                    code: 'ERROR_INVALID_DOMAIN',
                    cause: error,
                });
            }
            else if (publicKey.rpId !== effectiveDomain) {
                return new WebAuthnError({
                    message: `The RP ID "${publicKey.rpId}" is invalid for this domain`,
                    code: 'ERROR_INVALID_RP_ID',
                    cause: error,
                });
            }
        }
        else if (error.name === 'UnknownError') {
            return new WebAuthnError({
                message: 'The authenticator was unable to process the specified options, or could not create a new assertion signature',
                code: 'ERROR_AUTHENTICATOR_GENERAL_ERROR',
                cause: error,
            });
        }
        return error;
    }

    async function startAuthentication(requestOptionsJSON, useBrowserAutofill = false) {
        if (!browserSupportsWebAuthn()) {
            throw new Error('WebAuthn is not supported in this browser');
        }
        let allowCredentials;
        if (requestOptionsJSON.allowCredentials?.length !== 0) {
            allowCredentials = requestOptionsJSON.allowCredentials?.map(toPublicKeyCredentialDescriptor);
        }
        const publicKey = {
            ...requestOptionsJSON,
            challenge: base64URLStringToBuffer(requestOptionsJSON.challenge),
            allowCredentials,
        };
        const options = {};
        if (useBrowserAutofill) {
            if (!(await browserSupportsWebAuthnAutofill())) {
                throw Error('Browser does not support WebAuthn autofill');
            }
            const eligibleInputs = document.querySelectorAll('input[autocomplete*=\'webauthn\']');
            if (eligibleInputs.length < 1) {
                throw Error('No <input> with `"webauthn"` in its `autocomplete` attribute was detected');
            }
            options.mediation = 'conditional';
            publicKey.allowCredentials = [];
        }
        options.publicKey = publicKey;
        options.signal = webauthnAbortService.createNewAbortSignal();
        let credential;
        try {
            credential = (await navigator.credentials.get(options));
        }
        catch (err) {
            throw identifyAuthenticationError({ error: err, options });
        }
        if (!credential) {
            throw new Error('Authentication was not completed');
        }
        const { id, rawId, response, type } = credential;
        let userHandle = undefined;
        if (response.userHandle) {
            userHandle = bufferToUTF8String(response.userHandle);
        }
        return {
            id,
            rawId: bufferToBase64URLString(rawId),
            response: {
                authenticatorData: bufferToBase64URLString(response.authenticatorData),
                clientDataJSON: bufferToBase64URLString(response.clientDataJSON),
                signature: bufferToBase64URLString(response.signature),
                userHandle,
            },
            type,
            clientExtensionResults: credential.getClientExtensionResults(),
            authenticatorAttachment: toAuthenticatorAttachment(credential.authenticatorAttachment),
        };
    }

    /** This function begins the registration process
     *  @param {string} usernameEl - username of user to register
     *  @param {string} statusEl - status element
     */
    function register(usernameEl, statusEl, btnId) {
        return __awaiter(this, void 0, void 0, function* () {
            const usernameInput = document.getElementById(usernameEl);
            const statusLabel = document.getElementById(statusEl);
            document.getElementById(btnId);
            const resp = yield fetch(`/auth/register/begin/${usernameInput.value}`);
            if (!resp.ok) {
                clearClasslist([usernameInput, statusLabel]);
                usernameInput.classList.add("input-error");
                statusLabel.classList.add("text-error");
                statusLabel.innerHTML = yield resp.text();
                return;
            }
            const registrationOptions = (yield resp.json()).publicKey;
            let attResp;
            try {
                // Pass the options to the authenticator and wait for a response
                attResp = yield startRegistration(registrationOptions);
            }
            catch (error) {
                // Some basic error handling
                if (error.name === "InvalidStateError") {
                    console.log("Error: Authenticator was probably already registered by user");
                }
                else {
                    console.log(error);
                }
                throw error;
            }
            const result = yield fetch(`/auth/verify-registration/${usernameInput.value}`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(attResp),
            });
            if (!result.ok) {
                clearClasslist([usernameInput, statusLabel]);
                usernameInput.classList.add("input-error");
                statusLabel.classList.add("text-error");
                statusLabel.innerHTML = yield result.text();
            }
            else {
                clearClasslist([usernameInput, statusLabel]);
                usernameInput.classList.add("input-success");
                statusLabel.classList.add("text-success");
                statusLabel.innerHTML = "Success! You can now login.";
            }
        });
    }
    window.registerClick = register;
    function login(usernameEl, statusEl, btnId) {
        return __awaiter(this, void 0, void 0, function* () {
            const usernameInput = document.getElementById(usernameEl);
            const statusLabel = document.getElementById(statusEl);
            const btn = document.getElementById(btnId);
            const fetchLoginOptions = (username) => __awaiter(this, void 0, void 0, function* () { return fetch(`/auth/generate-authentication-options/${username}`); });
            const verifyLogin = (attResp, username) => __awaiter(this, void 0, void 0, function* () {
                return fetch(`/auth/verify-authentication/${username}`, {
                    method: "POST",
                    headers: {
                        "Content-Type": "application/json",
                    },
                    body: JSON.stringify(attResp),
                });
            });
            const resp = yield fetchLoginOptions(usernameInput.value);
            if (!resp.ok) {
                clearClasslist([usernameInput, statusLabel]);
                usernameInput.classList.add("input-error");
                statusLabel.classList.add("text-error");
                statusLabel.innerHTML = yield resp.text();
                return;
            }
            const options = (yield resp.json());
            let loginResp = yield startAuthentication(options.publicKey);
            let result = yield verifyLogin(loginResp, usernameInput.value);
            if (!result.ok) {
                clearClasslist([usernameInput, statusLabel]);
                usernameInput.classList.add("input-error");
                statusLabel.classList.add("text-error");
                statusLabel.innerHTML = yield result.text();
                return;
            }
            else {
                clearClasslist([usernameInput, statusLabel]);
                usernameInput.classList.add("input-success");
                statusLabel.classList.add("text-success");
                statusLabel.innerHTML = "Success! Redirecting...";
            }
            btn.dispatchEvent(new Event("refreshModal"));
        });
    }
    window.loginClick = login;
    function clearClasslist(elements) {
        elements.forEach((x) => (x.classList.value = x.classList.value
            .split(" ")
            .filter((y) => !y.includes("success") && !y.includes("error"))
            .join(" ")));
    }

})();