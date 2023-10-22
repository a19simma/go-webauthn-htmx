import {
  startAuthentication,
  startRegistration,
} from "@simplewebauthn/browser";
import {
  AuthenticationResponseJSON,
  PublicKeyCredentialRequestOptionsJSON,
} from "@simplewebauthn/typescript-types";

export {};
declare global {
  interface Window {
    registerClick: Function;
    loginClick: Function;
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
    clearClasslist([usernameInput, statusLabel]);
    usernameInput.classList.add("input-error");
    statusLabel.classList.add("text-error");
    statusLabel.innerHTML = await resp.text();
    return;
  }
  const registrationOptions = (await resp.json()).publicKey;
  let attResp;
  try {
    // Pass the options to the authenticator and wait for a response
    attResp = await startRegistration(registrationOptions);
  } catch (error: any) {
    // Some basic error handling
    if (error.name === "InvalidStateError") {
      console.log(
        "Error: Authenticator was probably already registered by user",
      );
    } else {
      console.log(error);
    }
    throw error;
  }
  const result = await fetch(
    `/auth/verify-registration/${usernameInput.value}`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(attResp),
    },
  );
  if (!result.ok) {
    clearClasslist([usernameInput, statusLabel]);
    usernameInput.classList.add("input-error");
    statusLabel.classList.add("text-error");
    statusLabel.innerHTML = await result.text();
  } else {
    clearClasslist([usernameInput, statusLabel]);
    usernameInput.classList.add("input-success");
    statusLabel.classList.add("text-success");
    statusLabel.innerHTML = "Success! You can now login.";
  }
}
window.registerClick = register;

async function login(usernameEl: string, statusEl: string, btnId: string) {
  const usernameInput = document.getElementById(usernameEl) as HTMLInputElement;
  const statusLabel = document.getElementById(statusEl) as HTMLElement;
  const btn = document.getElementById(btnId) as HTMLButtonElement;
  const fetchLoginOptions = async (username: string) =>
    fetch(`/auth/generate-authentication-options/${username}`);

  const verifyLogin = async (
    attResp: AuthenticationResponseJSON,
    username: string,
  ) =>
    fetch(`/auth/verify-authentication/${username}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(attResp),
    });
  const resp = await fetchLoginOptions(usernameInput.value);
  if (!resp.ok) {
    clearClasslist([usernameInput, statusLabel]);
    usernameInput.classList.add("input-error");
    statusLabel.classList.add("text-error");
    statusLabel.innerHTML = await resp.text();
    return;
  }
  const options = (await resp.json()) as any;
  let loginResp = await startAuthentication(options.publicKey);
  let result = await verifyLogin(loginResp, usernameInput.value);
  if (!result.ok) {
    clearClasslist([usernameInput, statusLabel]);
    usernameInput.classList.add("input-error");
    statusLabel.classList.add("text-error");
    statusLabel.innerHTML = await result.text();
    return;
  } else {
    clearClasslist([usernameInput, statusLabel]);
    usernameInput.classList.add("input-success");
    statusLabel.classList.add("text-success");
    statusLabel.innerHTML = "Success! Redirecting...";
  }
  btn.dispatchEvent(new Event("refreshModal"));
}
window.loginClick = login;

function clearClasslist(elements: HTMLElement[]) {
  elements.forEach(
    (x) =>
      (x.classList.value = x.classList.value
        .split(" ")
        .filter((y) => !y.includes("success") && !y.includes("error"))
        .join(" ")),
  );
}
