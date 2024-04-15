import { GetServerSideProps, InferGetServerSidePropsType } from 'next'
import { useRouter } from 'next/router'
import Link from 'next/link'
import { getConfig } from '@/lib/config'
import { Table, TableHeader, TableColumn, TableBody, TableRow, TableCell, getKeyValue, SortDescriptor } from '@nextui-org/react'
import React from 'react'
import { getTransactionsAndInnoDBStatus, ProcessWithTransaction } from '@/lib/transaction_repository'

type Repo = {
    processList: ProcessWithTransaction[]
    innodbStatus: string
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

export const getServerSideProps = (async (context) => {
    const instance = getConfig().instances[context.query.identifier as string]

    if (!instance) {
        return {
            redirect: {
                destination: '/',
                permanent: false,
            },
        }
    }

    const [processListWithTransaction, innoDbStatusString] = await getTransactionsAndInnoDBStatus(instance)

    const repo: Repo = {
        processList: processListWithTransaction,
        innodbStatus: innoDbStatusString,
    }
    // Pass data to the page via props
    return { props: { repo } }
}) satisfies GetServerSideProps<{ repo: Repo }>

export default function Home({ repo }: InferGetServerSidePropsType<typeof getServerSideProps>) {
    const router = useRouter()

    const [processList, setProcessList] = React.useState<ProcessWithTransaction[]>(repo.processList)

    const sort: (descriptor: SortDescriptor) => void = (descriptor) => {
        const key = descriptor.column as string
        const direction = descriptor.direction

        repo.processList.sort((a, b) => {
            const aValue = getKeyValue(a, key)
            const bValue = getKeyValue(b, key)

            if (direction === 'ascending') {
                return aValue > bValue ? 1 : -1
            } else {
                return aValue < bValue ? 1 : -1
            }
        })

        setProcessList([...repo.processList])
    }

    return (
        <main>
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
            <Table
                aria-label="Example table with client side sorting"
                sortDescriptor={{ column: 'id', direction: 'ascending' }}
                onSortChange={sort}
                isCompact={true}
                isStriped={true}
            >
                <TableHeader>
                    <TableColumn key="kill">🔥</TableColumn>
                    <TableColumn allowsSorting={true} key="id">
                        Id
                    </TableColumn>
                    <TableColumn allowsSorting={true} key="user">
                        User
                    </TableColumn>
                    <TableColumn allowsSorting={true} key="host">
                        Host
                    </TableColumn>
                    <TableColumn allowsSorting={true} key="db">
                        db
                    </TableColumn>
                    <TableColumn allowsSorting={true} key="command">
                        Command
                    </TableColumn>
                    <TableColumn allowsSorting={true} key="time">
                        Time
                    </TableColumn>
                    <TableColumn key="state">State</TableColumn>
                    <TableColumn allowsSorting={true} key="info">
                        Info
                    </TableColumn>
                    <TableColumn allowsSorting={true} key="progress">
                        Progress
                    </TableColumn>
                    <TableColumn allowsSorting={true} key="transactionTime">
                        Transaction Time
                    </TableColumn>
                    <TableColumn allowsSorting={true} key="transactionInfo">
                        Transaction Info
                    </TableColumn>
                </TableHeader>
                <TableBody items={processList}>
                    {(item) => (
                        <TableRow key={item.Id} className={`${item.transaction?.activeTime && item.transaction.activeTime > 10 ? 'bg-red-300' : ''}`}>
                            <TableCell className="align-top">
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
                            <TableCell className="align-top">{item.Id}</TableCell>
                            <TableCell className="align-top">
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
                            <TableCell className="align-top">{item.Host}</TableCell>
                            <TableCell className="align-top">
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
                            <TableCell className="align-top">{item.Command}</TableCell>
                            <TableCell className="align-top">{item.Time} s</TableCell>
                            <TableCell className="align-top">{item.State}</TableCell>
                            <TableCell className="align-top font-mono">{item.Info}</TableCell>
                            <TableCell className="align-top">{item.Progress}</TableCell>
                            <TableCell className="align-top">{item.transaction?.activeTime} s</TableCell>
                            <TableCell className="align-top font-mono whitespace-pre-line">{item.transaction?.info.join('\n')}</TableCell>
                        </TableRow>
                    )}
                </TableBody>
            </Table>
            <div className="w-9/12 m-5 collapse collapse-plus bg-base-200">
                <input type="checkbox" />
                <div className="collapse-title text-xl font-medium">Click to see complete innodb status result.</div>
                <div className="collapse-content whitespace-pre-line font-mono">{repo.innodbStatus}</div>
            </div>
        </main>
    )
}
