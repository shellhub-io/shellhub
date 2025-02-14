const displayOnlyTenCharacters = (str : string) => {
  if (str !== undefined) {
    if (str.length > 10) return `${str.slice(0, 10)}...`;
  }
  return str;
};

export default displayOnlyTenCharacters;
