import http from '@/store/helpers/http';

export const updateTag = async (data) => http().put(`tags/${data.oldTag}`, { tag: data.newTag });

export const removeTag = async (tag) => http().delete(`tags/${tag}`);

export const getTags = async () => http().get('/tags');
