import moment from "moment";

const formatDate = (date: string) => {
  if (date) {
    return moment(date).format("dddd, MMMM Do YYYY, h:mm:ss a");
  }
  return null;
};

const lastSeen = (date: string) => {
  if (date) {
    return moment(date).fromNow();
  }
  return null;
};

export { formatDate, lastSeen };
