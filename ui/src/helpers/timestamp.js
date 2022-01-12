export default (timestamp) => new Date(timestamp * 1000).toDateString().slice(4);
