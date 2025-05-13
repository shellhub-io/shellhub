const displayOnlyTenCharacters = (str: string) => str.length > 10 ? `${str.slice(0, 10)}...` : str;

export default displayOnlyTenCharacters;
