import moment from "moment";

export const formatDate = (date: string) => {
  if (date) {
    return moment(date).format("dddd, MMMM Do YYYY, h:mm:ss a");
  }
  return null;
};

export const formatDateCompact = (date: string) => {
  if (date) {
    return moment(date).format("LLL");
  }
  return null;
};

export const formatDateWithoutDayAndHours = (date : number) => {
  if (date) {
    return moment.unix(date).format("MMM Do YYYY");
  }
  return null;
};

export const formatDateFullAbrevied = (date : string) => {
  if (date) {
    return moment(date).format("ddd, MMM Do YY, h:mm:ss a");
  }
  return null;
};

export const lastSeen = (date: string) => {
  if (date) {
    return moment(date).fromNow();
  }
  return null;
};
