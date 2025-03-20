import moment from "moment";

const formatDate = (date: string, formatToken: string) => date ? moment(date).format(formatToken) : null;

export const formatFullDateTime = (date: string) => formatDate(date, "dddd, MMMM Do YYYY, h:mm:ss a");

export const formatShortDateTime = (date: string) => formatDate(date, "LLL");

export const formatAbbreviatedDateTime = (date : string) => formatDate(date, "ddd, MMM Do YY, h:mm:ss a");

export const formatUnixToDate = (timestamp: number) => timestamp ? moment.unix(timestamp).format("MMM Do YYYY") : null;

export const getTimeFromNow = (date: string) => date ? moment(date).fromNow() : null;
