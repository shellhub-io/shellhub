import moment from "moment";

/** Formats the date to "Saturday, March 1st 2025, 9:27:30 am" */
export const formatFullDateTime = (date: string) => date ? moment(date).format("dddd, MMMM Do YYYY, h:mm:ss a") : null;

/** Returns the time difference from the given date (e.g. "3 hours ago") */
export const getTimeFromNow = (date: string) => date ? moment(date).fromNow() : null;
