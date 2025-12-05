create table message_updates (
    id           bigint primary key,
    message      JSON not null default '{}'::JSON,
    is_processed boolean not null default false,
    created_at   timestamp not null default now()
);

comment on column message_updates.id IS 'Поле update_id сообщения Телеграм';
comment on column message_updates.message IS 'JSON данные сообщения';
comment on column message_updates.is_processed IS 'Признак обработки сообщения';
comment on column message_updates.created_at IS 'Дата и время получения сообщения';
