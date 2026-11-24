const LETTERS_AND_NUMBERS = /^[a-zA-Z0-9]*$/;

/**
 * What keeps a tag name from being well-formed, or null when it is: 3 to 255 characters, letters
 * and numbers only, the format the API checks. hint is the text to show the user as is. Length is
 * checked first, so any name under 3 characters, the empty name included, gets "At least 3
 * characters"; incomplete is true only when it also holds nothing but letters and numbers, so
 * typing more can make it well-formed. The caller trims the name. A well-formed name can still be
 * refused by the API, as a duplicate or a tag over a device's limit.
 */
export function tagNameProblem(
  name: string,
): { hint: string; incomplete: boolean } | null {
  if (name.length < 3) {
    return {
      hint: "At least 3 characters",
      incomplete: LETTERS_AND_NUMBERS.test(name),
    };
  }
  if (name.length > 255) {
    return { hint: "At most 255 characters", incomplete: false };
  }
  if (!LETTERS_AND_NUMBERS.test(name)) {
    return { hint: "Only letters and numbers", incomplete: false };
  }
  return null;
}
