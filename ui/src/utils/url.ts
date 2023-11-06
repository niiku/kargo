export const urlWithProtocol = (url: string) => url == undefined || null ? '' : `https://${url.replace('https://', '')}`;
