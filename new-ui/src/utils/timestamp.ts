export default (timestamp: number) =>
  new Date(timestamp * 1000).toDateString().slice(4);
