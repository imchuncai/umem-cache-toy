.. SPDX-License-Identifier: GPL-2.0-only
.. Copyright (C) 2026, Shu De Zheng <imchuncai@gmail.com>. All Rights Reserved.

==============
Umem-cache-toy
==============

Umem-cache-toy is a toy CLI for experimenting with
`Umem-cache <https://github.com/imchuncai/umem-cache>`_.

Usage
=====
::

	go run . -h

	Usage:
	  umem-cache-toy [-tls_key <path> -tls_cert <path> -tls_ca_cert <path>]
	                 [-timeout <seconds>] <address...>

	Example:
	  umem-cache-toy [::1]:10047
