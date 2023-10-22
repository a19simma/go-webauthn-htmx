import { startAuthentication, startRegistration } from "@simplewebauthn/browser";
import { AuthenticationResponseJSON } from '@simplewebauthn/typescript-types';

export { };
declare global {
    interface Window {
        registerClick: Function
        loginClick: Function
    }
}
/** This function begins the registration process
 *  @param {string} usernameEl - username of user to register
 *  @param {string} statusEl - status element
 */
async function register(usernameEl: string, statusEl: string, btnId: string) {
    const usernameInput = document.getElementById(usernameEl) as HTMLInputElement;
    const statusLabel = document.getElementById(statusEl) as HTMLElement;
    const btn = document.getElementById(btnId) as HTMLButtonElement;
    const resp = await fetch(`/auth/register/begin/${usernameInput.value}`);
    if (!resp.ok) {
        return resp;
    }
    const registrationOptions = (await resp.json()).publicKey;
    let attResp;
    try {
        // Pass the options to the authenticator and wait for a response
        attResp = await startRegistration(registrationOptions);
    }
    catch (error: any) {
        // Some basic error handling
        if (error.name === "InvalidStateError") {
            console.log("Error: Authenticator was probably already registered by user");
        }
        else {
            console.log(error);
        }
        throw error;
    }
    const result = await fetch(`/auth/verify-registration/${usernameInput.value}`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify(attResp),
    });
    if (!result.ok) {
        usernameInput.classList.add("input-error");
        statusLabel.classList.add("text-error");
        statusLabel.innerHTML = await result.text();
    }
    else {
        usernameInput.classList.add("input-success");
        statusLabel.classList.add("text-success");
        statusLabel.innerHTML = "Success! You can now login.";
    }
    btn.dispatchEvent(new Event("refreshModal"));
}
window.registerClick = register;

async function login(usernameEl: string, statusEl: string, btnId: string) {
    const usernameInput = document.getElementById(usernameEl) as HTMLInputElement;
    const statusLabel = document.getElementById(statusEl) as HTMLElement;
    const btn = document.getElementById(btnId) as HTMLButtonElement;
    const fetchLoginOptions = async (username: string) =>
        fetch(`/auth/generate-authentication-options/${username}`).then((resp) => resp.json());

    const verifyLogin = async (attResp: AuthenticationResponseJSON, username: string) =>
        fetch(`/auth/verify-authentication/${username}`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(attResp),
        });
    const options = await fetchLoginOptions(usernameInput.value);
    let loginResp = await startAuthentication(options.publicKey);
    let result = await verifyLogin(loginResp, usernameInput.value);
    if (!result.ok) {
        usernameInput.classList.add("input-error");
        statusLabel.classList.add("text-error");
        statusLabel.innerHTML = await result.text();
    }
    else {
        usernameInput.classList.add("input-success");
        statusLabel.classList.add("text-success");
        statusLabel.innerHTML = "Success! You are now logged in.";
    }
    btn.dispatchEvent(new Event("refreshModal"));
}
window.loginClick = login;

