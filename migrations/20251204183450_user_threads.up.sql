create table user_threads (
    user_id bigint primary key,
    thread_id bigint not null
);

comment on column user_threads.user_id IS 'ID пользователя';
comment on column user_threads.thread_id IS 'ID треда пользователя';
