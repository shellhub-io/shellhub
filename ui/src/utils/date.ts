import moment from "moment";

const formatDate = (date: string, formatToken: string) => date ? moment(date).format(formatToken) : null;

/** Formats the date to "Saturday, March 1st 2025, 9:27:30 am" */
export const formatFullDateTime = (date: string) => formatDate(date, "dddd, MMMM Do YYYY, h:mm:ss a");

/** Formats the date to "March 1st 2025, 9:27:30 am" */
export const formatShortDateTime = (date: string) => formatDate(date, "LLL");

/** Formats the date to "Sat, Mar 1st 25, 9:27:30 am" */
export const formatAbbreviatedDateTime = (date: string) => formatDate(date, "ddd, MMM Do YY, h:mm:ss a");

/** Formats the date (an Unix timestamp) to "Mar 1st 2025" */
export const formatUnixToDate = (timestamp: number) => timestamp ? moment.unix(timestamp).format("MMM Do YYYY") : null;

/** Returns the time difference from the given date (e.g. "3 hours ago") */
export const getTimeFromNow = (date: string) => date ? moment(date).fromNow() : null;
