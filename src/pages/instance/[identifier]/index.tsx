import { useRouter } from 'next/router'
import Link from 'next/link'
import { Table, TableHeader, TableColumn, TableBody, TableRow, TableCell, getKeyValue, Spinner } from '@nextui-org/react'
import React, { useEffect } from 'react'
import { ProcessWithTransaction } from '@/lib/transaction_repository'
import { useAsyncList } from '@react-stately/data'

const formatNumber = function (num: number | null | undefined): string {
    if (num === null || num === undefined) {
        return ''
    }
    return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ' ')
}

const stringToColor = function (str: string | null): string {
    if (!str) {
        return '#000000'
    }
    let hash = 0
    for (let i = 0; i < str.length; i++) {
        hash = str.charCodeAt(i) + ((hash << 5) - hash)
    }
    let colour = '#'
    for (let i = 0; i < 3; i++) {
        let value = (hash >> (i * 8)) & 0xff
        colour += ('00' + value.toString(16)).slice(-2)
    }
    return colour
}

const blackOrWhite = function (hex: string): string {
    const r = parseInt(hex.slice(1, 2), 16)
    const g = parseInt(hex.slice(3, 2), 16)
    const b = parseInt(hex.slice(5, 2), 16)
    const brightness = (r * 299 + g * 587 + b * 114) / 1000
    return brightness > 155 ? '#000000' : '#ffffff'
}

export default function ShowInstance() {
    const router = useRouter()

    const [isLoading, setIsLoading] = React.useState(true)
    const [innodbStatus, setInnodbStatus] = React.useState<string>('')

    let list = useAsyncList<ProcessWithTransaction>({
        async load({ signal }) {
            const instance = router.query.identifier as string
            if (!instance) {
                return {
                    items: [],
                }
            }
            let res = await fetch(`/api/processlist?instance=${instance}`, {
                signal,
            })
            let json = await res.json()
            setIsLoading(false)
            setInnodbStatus(json.innoDbStatus)

            return {
                items: json.processList,
            }
        },
        async sort({ items, sortDescriptor }) {
            const key = sortDescriptor.column as string
            return {
                items: items.sort((a, b) => {
                    let first = getKeyValue(a, key)
                    let second = getKeyValue(b, key)
                    if (key === 'transactionTime') {
                        first = a.transaction?.activeTime
                        second = b.transaction?.activeTime
                    }

                    if (key === 'transactionInfo') {
                        first = a.transaction?.info.join('\n')
                        second = b.transaction?.info.join('\n')
                    }

                    let cmp

                    if (first == null && second != null) {
                        cmp = -1
                    } else if (first != null && second == null) {
                        cmp = 1
                    } else {
                        cmp = (parseInt(first) || first) < (parseInt(second) || second) ? -1 : 1
                    }

                    if (sortDescriptor.direction === 'descending') {
                        cmp *= -1
                    }

                    return cmp
                }),
            }
        },
    })

    useEffect(() => {
        list.reload()
    }, [router.query.identifier])

    return (
        <main className={'overflow-x-scroll'}>
            <div className="text-xl breadcrumbs">
                <ul>
                    <li>
                        <Link href="/">Instance Selector</Link>
                    </li>
                    <li>
                        <a>{router.query.identifier}</a>
                    </li>
                </ul>
            </div>
            <Table aria-label="Example table with client side sorting" sortDescriptor={list.sortDescriptor} onSortChange={list.sort} isCompact={true}>
                <TableHeader>
                    <TableColumn key="kill">🔥</TableColumn>
                    <TableColumn allowsSorting={true} key="Id">
                        Id
                    </TableColumn>
                    <TableColumn allowsSorting={true} key="User">
                        User
                    </TableColumn>
                    <TableColumn allowsSorting={true} key="Host">
                        Host
                    </TableColumn>
                    <TableColumn allowsSorting={true} key="db">
                        db
                    </TableColumn>
                    <TableColumn allowsSorting={true} key="Command">
                        Command
                    </TableColumn>
                    <TableColumn allowsSorting={true} key="Time">
                        Time
                    </TableColumn>
                    <TableColumn allowsSorting={true} key="State">
                        State
                    </TableColumn>
                    <TableColumn allowsSorting={true} key="Info">
                        Info
                    </TableColumn>
                    <TableColumn allowsSorting={true} key="Progress">
                        Progress
                    </TableColumn>
                    <TableColumn allowsSorting={true} key="transactionTime">
                        Transaction Time
                    </TableColumn>
                    <TableColumn allowsSorting={true} key="transactionInfo">
                        Transaction Info
                    </TableColumn>
                </TableHeader>
                <TableBody items={list.items} isLoading={isLoading} loadingContent={<Spinner label="Loading..." />}>
                    {(item) => (
                        <TableRow key={item.Id} className={`${item.transaction?.activeTime && item.transaction.activeTime > 10 ? 'bg-red-300' : ''}`}>
                            <TableCell className="align-top text-nowrap">
                                <button
                                    onClick={async () => {
                                        if (!confirm(`Are you sure you want to kill process ${item.Id} by user '${item.User}'?`)) {
                                            return
                                        }
                                        await fetch('/api/kill', {
                                            method: 'POST',
                                            headers: {
                                                'Content-Type': 'application/json',
                                            },
                                            body: JSON.stringify({
                                                id: item.Id,
                                                instance: router.query.identifier as string,
                                            }),
                                        }).then(() => {
                                            // Refresh the page after the request is done
                                            window.location.reload()
                                        })
                                    }}
                                >
                                    💀
                                </button>
                            </TableCell>
                            <TableCell className="align-top text-nowrap">{item.Id}</TableCell>
                            <TableCell className="align-top text-nowrap">
                                <div
                                    style={{
                                        color: blackOrWhite(stringToColor(item.User)),
                                        backgroundColor: stringToColor(item.User),
                                        borderColor: stringToColor(item.User),
                                    }}
                                    className={'badge'}
                                >
                                    {item.User}
                                </div>
                            </TableCell>
                            <TableCell className="align-top text-nowrap">{item.Host}</TableCell>
                            <TableCell className="align-top text-nowrap">
                                <div
                                    style={{
                                        color: blackOrWhite(stringToColor(item.db)),
                                        backgroundColor: stringToColor(item.db),
                                        borderColor: stringToColor(item.db),
                                    }}
                                    className={'badge'}
                                >
                                    {item.db}
                                </div>
                            </TableCell>
                            <TableCell className="align-top text-nowrap">{item.Command}</TableCell>
                            <TableCell className="align-top text-nowrap">{formatNumber(item.Time)} s</TableCell>
                            <TableCell className="align-top text-nowrap">{item.State}</TableCell>
                            <TableCell className="align-top font-mono text-nowrap">{item.Info}</TableCell>
                            <TableCell className="align-top text-nowrap">{item.Progress}</TableCell>
                            <TableCell className="align-top text-nowrap">
                                {formatNumber(item.transaction?.activeTime)} {item.transaction?.activeTime ? 's' : ''}
                            </TableCell>
                            <TableCell className="align-top font-mono whitespace-pre-line text-nowrap">{item.transaction?.info.join('\n')}</TableCell>
                        </TableRow>
                    )}
                </TableBody>
            </Table>
            <div className="w-9/12 m-5 collapse collapse-plus bg-base-200">
                <input type="checkbox" />
                <div className="collapse-title text-xl font-medium">Click to see complete innodb status result.</div>
                <div className="collapse-content whitespace-pre-line font-mono">{innodbStatus}</div>
            </div>
        </main>
    )
}
