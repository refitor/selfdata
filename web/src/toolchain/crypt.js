import {Base64} from 'js-base64'

export function EncodeByBase64(data) {
  return Base64.encode(data)
}
export function DecodeByBase64(data) {
  return Base64.decode(data)
}

export function EncodeURIByBase64(data) {
  return Base64.encodeURI(data)
}
export function DecodeURIByBase64(data) {
  return Base64.decode(data)
}

