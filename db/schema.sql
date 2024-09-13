create table employee
(
    id       uuid primary key,
    username text unique not null
);

create table organization
(
    id   uuid primary key,
    name text not null
);

create table organization_employee
(
    organization_id uuid references organization (id) not null,
    employee_id     uuid references employee (id)     not null,
    unique (organization_id, employee_id)
);

create type tender_status as enum ('Created', 'Published', 'Closed');

create type service_type as enum ('Construction', 'Delivery', 'Manufacture');

create table tender
(
    id              uuid primary key,
    name            text                              not null,
    description     text                              not null,
    status          tender_status                     not null,
    service_type    service_type                      not null,
    organization_id uuid references organization (id) not null,
    creator_id      uuid references employee (id)     not null,
    version_id      bigint                            not null,
    created         timestamp                         not null
);

create table tender_version
(
    id           bigint,
    tender_id    uuid references tender (id) not null,
    name         text                        not null,
    description  text                        not null,
    status       tender_status               not null,
    service_type service_type                not null,
    created      timestamp                   not null,
    unique (tender_id, id)
);

create type bid_status as enum ('Created', 'Published', 'Canceled', 'Approved', 'Rejected');

create type creator_type as enum ('Organization', 'User');

create table bid
(
    id              uuid primary key,
    name            text                              not null,
    description     text                              not null,
    status          bid_status                        not null,
    tender_id       uuid references tender (id)       not null,
    creator_type    creator_type                      not null,
    creator_id      uuid references employee (id)     not null,
    organization_id uuid references organization (id) not null,
    version_id      bigint                            not null,
    created         timestamp                         not null
);

create table bid_version
(
    id          bigint                   not null,
    bid_id      uuid references bid (id) not null,
    name        text                     not null,
    description text                     not null,
    status      bid_status               not null,
    created_    timestamp                not null,
    unique (bid_id, id)
);

create table bid_agreement
(
    bid_id      uuid references bid (id)      not null,
    employee_id uuid references employee (id) not null,
    status      bid_status                    not null,
    unique (bid_id, employee_id)
);
