const showTag = (str: string) => {
  if (str !== undefined) {
    if (str.length > 10) {
      return true;
    }
  }
  return false;
};

export default showTag;
